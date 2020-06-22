// Package internal poll camera images
package internal

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/publisher"
)

// readCameraImage returns the image downloaded from the URL and the duration to download it.
func (ipcam *IPCamApp) readImage(url string, login string, password string) ([]byte, time.Duration, error) {
	logger := ipcam.logger
	logger.Debugf("readCameraImage: Reading camera image from URL %s", url)
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
		logger.Errorf("readCameraImage: Error opening URL %s: %s", url, err)
		return nil, 0, err
	}
	defer resp.Body.Close()
	// was it a good response?
	if resp.StatusCode > 299 {
		logger.Errorf("readCameraImage: Failed opening URL %s: %s", url, resp.Status)
		err = errors.New(resp.Status)
		return nil, 0, err
	}
	image, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("readCameraImage: Error reading camera image from %s: %s", url, err)
		return nil, 0, err
	}
	endTime := time.Now()
	duration := endTime.Sub(startTime).Round(time.Millisecond)
	return image, duration, nil
}

// saveImage save image to file
func (ipcam *IPCamApp) saveImage(filePath string, image []byte) error {
	ipcam.logger.Debugf("saveImage: Saving image to file %s", filePath)
	err := ioutil.WriteFile(filePath, image, 0644)
	return err
}

// PollCamera and publish image.
// If the camera device doesn't have an image output it is added.
// If a filename property is configured then save the camera image to the file
// Images are publised as a binary array and are unsigned
func (ipcam *IPCamApp) PollCamera(camera *iotc.NodeDiscoveryMessage) ([]byte, error) {
	var err error
	cameras := ipcam.pub.Nodes
	camAddr := camera.Address
	url, _ := cameras.GetNodeConfigString(camAddr, iotc.NodeAttrURL, "")
	loginName, _ := cameras.GetNodeConfigString(camAddr, iotc.NodeAttrLoginName, "")
	password, _ := cameras.GetNodeConfigString(camAddr, iotc.NodeAttrPassword, "")
	ipcam.logger.Infof("pollCamera: Polling Camera %s image from %s", camera.NodeID, url)

	image, latency, err := ipcam.readImage(url, loginName, password)
	latencyStr := latency.String()

	if image != nil {
		//latency3 := math.Round(latency.Seconds()*1000)/1000
		cameras.SetNodeStatus(camAddr, iotc.NodeStatusMap{iotc.NodeStatusLatencyMSec: latencyStr})
		ipcam.pub.UpdateOutputValue(camera.NodeID, iotc.OutputTypeLatency, iotc.DefaultOutputInstance, latencyStr)

		// if a filename attribute is defined, save the image to the file
		// this is a bit of an oddball as this is for local use only, yet a published attribute?
		filename := cameras.GetNodeAttr(camAddr, iotc.NodeAttrFilename)
		if filename != "" {
			err = ipcam.saveImage(filename, image)
		}
		//
		output := ipcam.pub.GetOutputByType(camera.NodeID, iotc.OutputTypeImage, iotc.DefaultOutputInstance)
		// Dont store the image in memory as it consumes memory unnecesary
		// Don't sign so the image is directly usable by 3rd party (todo: add signing as config)
		ipcam.pub.PublishRaw(output, false, image)

		ipcam.pub.SetNodeErrorStatus(camera.NodeID, iotc.NodeRunStateReady, "")
	} else {
		// failed to get image from camera
		msg := fmt.Sprintf("Unable to get image from camera %s: %s", camera.NodeID, err)
		ipcam.pub.SetNodeErrorStatus(camera.NodeID, iotc.NodeRunStateError, msg)
		err = errors.New(msg)
	}
	return image, err
}

// Poll handler called on a 1 second interval. Each camera node has its own interval.
func (ipcam *IPCamApp) Poll(pub *publisher.Publisher) {
	// Each second check which cameras need to be polled and poll
	cameraList := pub.Nodes
	for _, camera := range pub.Nodes.GetAllNodes() {

		// Each camera can have its own poll interval. The fallback value is the ipcam poll interval
		pollDelay, _ := ipcam.pollDelay[camera.Address]
		if pollDelay <= 0 {
			pollInterval, _ := cameraList.GetNodeConfigInt(camera.Address, iotc.NodeAttrPollInterval, 600)
			pollDelay = pollInterval
			ipcam.logger.Debugf("pollLoop: Polling camera %s at interval of %d seconds", camera.NodeID, pollInterval)
			go ipcam.PollCamera(camera)
		}
		ipcam.pollDelay[camera.Address] = pollDelay - 1
	}
}
