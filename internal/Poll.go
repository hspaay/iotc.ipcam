// Package internal poll camera images
package internal

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	"github.com/iotdomain/iotdomain-go/publisher"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// readCameraImage returns the image downloaded from the URL and the duration to download it.
func (ipcam *IPCamApp) readImage(url string, login string, password string) ([]byte, time.Duration, error) {
	logrus.Debugf("readCameraImage: Reading camera image from URL %s", url)
	startTime := time.Now()
	var req *http.Request
	var resp *http.Response
	var err error

	if login == "" {
		// No auth
		resp, err = http.Get(url)
	} else {
		// basic auth
		client := &http.Client{}
		req, err = http.NewRequest("GET", url, nil)
		if err == nil {
			req.SetBasicAuth(login, password)
			resp, err = client.Do(req)
		}
	}
	// handle failure to load the image
	if err != nil {
		logrus.Errorf("readCameraImage: Error opening URL %s: %s", url, err)
		return nil, 0, err
	}
	defer resp.Body.Close()
	// was it a good response?
	if resp.StatusCode > 299 {
		logrus.Errorf("readCameraImage: Failed opening URL %s: %s", url, resp.Status)
		err = errors.New(resp.Status)
		return nil, 0, err
	}
	image, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("readCameraImage: Error reading camera image from %s: %s", url, err)
		return nil, 0, err
	}
	endTime := time.Now()
	duration := endTime.Sub(startTime).Round(time.Millisecond)
	return image, duration, nil
}

// not using the library's input from http yet
func (ipcam *IPCamApp) handleInputMessage(input *types.InputDiscoveryMessage, sender string, value string) {
	logrus.Infof("handleInputMessage")
}

// saveImage save image to the image folder
func (ipcam *IPCamApp) saveImage(filename string, image []byte) error {
	filePath := path.Join(ipcam.config.ImageFolder, filename)
	logrus.Debugf("saveImage: Saving image to file %s", filePath)
	err := ioutil.WriteFile(filePath, image, 0644)
	return err
}

// PollCamera and publish image.
// If the camera doesn't have an image output it is added.
// If a filename property is configured then save the camera image to the file
// Images are publised as a binary array and are unsigned
func (ipcam *IPCamApp) PollCamera(camera *types.NodeDiscoveryMessage) ([]byte, error) {
	var err error
	pub := ipcam.pub
	url, _ := pub.GetNodeConfigString(camera.NodeID, types.NodeAttrURL, "")
	loginName, _ := pub.GetNodeConfigString(camera.NodeID, types.NodeAttrLoginName, "")
	password, _ := pub.GetNodeConfigString(camera.NodeID, types.NodeAttrPassword, "")
	logrus.Infof("pollCamera: Polling Camera %s image from %s", camera.NodeID, url)

	image, latency, err := ipcam.readImage(url, loginName, password)
	latencyStr := latency.String()

	if image != nil {
		//latency3 := math.Round(latency.Seconds()*1000)/1000
		pub.UpdateNodeStatus(camera.NodeID, types.NodeStatusMap{types.NodeStatusLatencyMSec: latencyStr})
		pub.UpdateOutputValue(camera.NodeID, types.OutputTypeLatency, types.DefaultOutputInstance, latencyStr)

		// if a filename attribute is defined, save the image to the file
		// this is a bit of an oddball as this is for local use only, yet a published attribute?
		filename := pub.GetNodeAttr(camera.NodeID, types.NodeAttrFilename)
		if filename != "" {
			err = ipcam.saveImage(filename, image)
		}
		//
		output := pub.GetOutputByNodeHWID(camera.NodeID, types.OutputTypeImage, types.DefaultOutputInstance)
		// Dont store the image in memory as it consumes memory unnecesary
		// Don't sign so the image is directly usable by 3rd party (todo: add signing as config)
		pub.PublishRaw(output, false, string(image))

		pub.UpdateNodeErrorStatus(camera.NodeID, types.NodeRunStateReady, "")
	} else {
		// failed to get image from camera
		msg := fmt.Sprintf("Unable to get image from camera %s: %s", camera.NodeID, err)
		pub.UpdateNodeErrorStatus(camera.NodeID, types.NodeRunStateError, msg)
		err = errors.New(msg)
	}
	return image, err
}

// Poll handler called on a 1 second interval. Each camera node has its own interval.
func (ipcam *IPCamApp) Poll(pub *publisher.Publisher) {
	// Each second check which cameras need to be polled and poll
	cameraList := pub.GetNodes()
	for _, camera := range cameraList {

		// Each camera can have its own poll interval. The fallback value is the ipcam poll interval
		pollDelay, _ := ipcam.pollDelay[camera.Address]
		if pollDelay <= 0 {
			pollInterval, _ := pub.GetNodeConfigInt(camera.NodeID, types.NodeAttrPollInterval, 600)
			pollDelay = pollInterval
			logrus.Debugf("pollLoop: Polling camera %s at interval of %d seconds", camera.NodeID, pollInterval)
			go ipcam.PollCamera(camera)
		}
		ipcam.pollDelay[camera.Address] = pollDelay - 1
	}
}
