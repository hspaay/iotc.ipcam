package internal

import (
	"os"
	"testing"
	"time"

	"github.com/iotdomain/iotdomain-go/publisher"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const configFolder = "../test"

//const TmpKelowna1Jpg = "/tmp/kelowna1.jpg"
const cam1Id = "Snowshed-east"
const cam2Id = "LaSilla"
const cam3Id = "Kelowna"
const cam3File = "../test/kelowna-snapshot.jpg"

var appConfig *IPCamConfig = &IPCamConfig{}

func TestLoadConfig(t *testing.T) {
	pub, err := publisher.NewAppPublisher(AppID, configFolder, appConfig, false)
	assert.NoError(t, err)
	app := NewIPCamApp(appConfig, pub)

	// the snowshed camera has an output with image sensor configured
	camera := pub.GetNodeByDeviceID(cam1Id)
	require.NotNil(t, camera)
	assert.Equal(t, cam1Id, camera.NodeID, "Incorrect name for camera")

	output := pub.GetOutputByDevice(cam1Id, types.OutputTypeImage, types.DefaultOutputInstance)
	require.NotNil(t, output, "Missing output for camera image")

	// for coverage ...
	app.HandleConfigCommand("someaddress", nil)
}

// TestReadCamera test reading camera image from remote location
func TestReadCamera(t *testing.T) {
	pub, _ := publisher.NewAppPublisher(AppID, configFolder, appConfig, false)
	ipcam := NewIPCamApp(appConfig, pub)

	camURL, _ := pub.GetNodeConfigString(cam1Id, types.NodeAttrURL, "") //
	loginName, _ := pub.GetNodeConfigString(cam1Id, types.NodeAttrLoginName, "")
	password, _ := pub.GetNodeConfigString(cam1Id, types.NodeAttrPassword, "")
	image, duration, err := ipcam.readImage(camURL, loginName, password)
	assert.NoError(t, err)
	assert.NotNil(t, image)
	assert.NotEqualf(t, 0, duration, "Expected duration > 0")
}

// TestPollCamera which polls the first camera image in the config and publishes the result
func TestPollCamera(t *testing.T) {
	pub, err := publisher.NewAppPublisher(AppID, configFolder, appConfig, false)
	assert.NoError(t, err, "Failed to create ipcam publisher")
	ipcam := NewIPCamApp(appConfig, pub)

	os.Remove(cam3File)
	pub.Start()

	camera := pub.GetNodeByDeviceID(cam1Id)
	assert.NotNil(t, camera) // camera node has to exist
	// the snowshed camera has an image sensor configured, required for publishing the camera during poll
	image, err := ipcam.PollCamera(camera)
	assert.NotNil(t, image)
	assert.NoError(t, err)

	// after polling the camera, its latency attribute must have been updated
	latencyValue := pub.GetOutputValueByDevice(
		camera.NodeID, types.OutputTypeLatency, types.DefaultOutputInstance)
	if assert.NotNil(t, latencyValue, "No output value for latency on node %s", camera.Address) {
		assert.NotZero(t, latencyValue.Value, "No latency in polling camera")
	}

	output := pub.GetOutputByDevice(cam1Id, types.OutputTypeImage, types.DefaultOutputInstance)
	assert.NotNil(t, output) // camera node has to exist
	assert.Equal(t, types.OutputTypeImage, output.OutputType, "Incorrect camera output type")

	// can take a second before files are written from the heartbeat loop
	time.Sleep(time.Second)
	assert.FileExists(t, cam3File, "Image not saved to file %s", cam3File)

	// TODO listen for topic?
	pub.Stop()

	// error case - poll invalid url
	pub.UpdateNodeConfigValues(camera.DeviceID, types.NodeAttrMap{
		types.NodeAttrURL: "http://localhost/badurl.jpg",
	})
	_, err = ipcam.PollCamera(camera)
	assert.Error(t, err)
}

// Test updating of the poll rate on cam2
func TestConfigPollRate(t *testing.T) {
	pub, _ := publisher.NewAppPublisher(AppID, configFolder, appConfig, false)
	_ = NewIPCamApp(appConfig, pub)
	pub.Start()

	cam1 := pub.GetNodeByDeviceID(cam1Id)
	pub.UpdateNodeConfigValues(cam1.NodeID,
		types.NodeAttrMap{types.NodeAttrPollInterval: "654"},
	)

	pollInterval, err := pub.GetNodeConfigInt(cam1Id, types.NodeAttrPollInterval, 612)
	assert.NoErrorf(t, err, "Poll interval config not found")
	assert.Equal(t, 654, pollInterval)
	time.Sleep(1 * time.Second)

	pub.UpdateNodeConfigValues(cam1.NodeID,
		types.NodeAttrMap{types.NodeAttrPollInterval: "33"},
	)
	time.Sleep(1 * time.Second)
	pollInterval2, _ := pub.GetNodeConfigInt(cam1Id, types.NodeAttrPollInterval, 600)
	assert.Equal(t, 33, pollInterval2)

	pub.Stop()
}

// TestStartStop of the ipcam
func TestStartStop(t *testing.T) {
	pub, _ := publisher.NewAppPublisher(AppID, configFolder, appConfig, false)
	NewIPCamApp(appConfig, pub)

	pub.Start()

	camera := pub.GetNodeByDeviceID(cam3Id)
	assert.NotNil(t, camera) // camera node has to exist
	time.Sleep(10 * time.Second)
	pub.Stop()
}
