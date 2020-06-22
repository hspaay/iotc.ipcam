package internal

import (
	"testing"
	"time"

	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/publisher"
	"github.com/stretchr/testify/assert"
)

const cacheFolder = "../test/cache"
const configFolder = "../test"

//const TmpKelowna1Jpg = "/tmp/kelowna1.jpg"
const cam1Id = "Snowshed-east"
const cam2Id = "LaSilla"
const cam3Id = "Kelowna"

var appConfig *IPCamConfig = &IPCamConfig{}

func TestLoadConfig(t *testing.T) {
	pub, err := publisher.NewAppPublisher(AppID, configFolder, cacheFolder, appConfig, false)
	assert.NoError(t, err)
	NewIPCamApp(appConfig, pub)

	// the snowshed camera has an output with image sensor configured
	camera := pub.GetNodeByID(cam1Id)
	if assert.NotNil(t, camera) { // camera node has to exist
		assert.Equal(t, "snowshed-east", camera.NodeID, "Incorrect name for camera")
	}
	output := pub.GetOutputByType(cam1Id, iotc.OutputTypeImage, iotc.DefaultOutputInstance)
	assert.NotNil(t, output, "Missing output for camera image")
}

// TestReadCamera test reading camera image from remote location
func TestReadCamera(t *testing.T) {
	pub, _ := publisher.NewAppPublisher(AppID, configFolder, cacheFolder, appConfig, false)
	ipcam := NewIPCamApp(appConfig, pub)

	camURL, _ := pub.GetNodeConfigString(cam1Id, iotc.NodeAttrURL, "") //
	loginName, _ := pub.GetNodeConfigString(cam1Id, iotc.NodeAttrLoginName, "")
	password, _ := pub.GetNodeConfigString(cam1Id, iotc.NodeAttrPassword, "")
	image, duration, err := ipcam.readImage(camURL, loginName, password)
	assert.NoError(t, err)
	assert.NotNil(t, image)
	assert.NotEqualf(t, 0, duration, "Expected duration > 0")
}

// TestPollCamera which polls the first camera image in the config and publishes the result
func TestPollCamera(t *testing.T) {
	pub, err := publisher.NewAppPublisher(AppID, configFolder, cacheFolder, appConfig, false)
	assert.NoError(t, err, "Failed to create ipcam publisher")
	ipcam := NewIPCamApp(appConfig, pub)

	pub.Start()

	camera := pub.GetNodeByID(cam1Id)
	assert.NotNil(t, camera) // camera node has to exist
	// the snowshed camera has an image sensor configured, required for publishing the camera during poll
	image, err := ipcam.PollCamera(camera)
	assert.NotNil(t, image)
	assert.NoError(t, err)

	// after polling the camera, its latency attribute must have been updated
	latencyValue := pub.OutputValues.GetOutputValueByType(
		camera, iotc.OutputTypeLatency, iotc.DefaultOutputInstance)
	if assert.NotNil(t, latencyValue, "No output value for latency on node %s", camera.Address) {
		assert.NotZero(t, latencyValue.Value, "No latency in polling camera")
	}

	output := pub.GetOutputByType(cam1Id, iotc.OutputTypeImage, iotc.DefaultOutputInstance)
	assert.NotNil(t, output) // camera node has to exist
	assert.Equal(t, iotc.OutputTypeImage, output.OutputType, "Incorrect camera output type")

	// TODO listen for topic
	pub.Stop()
}

// Test updating of the poll rate on cam2
func TestConfigPollRate(t *testing.T) {
	pub, _ := publisher.NewAppPublisher(AppID, configFolder, cacheFolder, appConfig, false)
	_ = NewIPCamApp(appConfig, pub)
	pub.Start()

	cam1 := pub.GetNodeByID(cam1Id)
	pub.Nodes.SetNodeConfigValues(cam1.Address,
		iotc.NodeAttrMap{iotc.NodeAttrPollInterval: "654"},
	)

	pollInterval, err := pub.GetNodeConfigInt(cam1Id, iotc.NodeAttrPollInterval, 612)
	assert.NoErrorf(t, err, "Poll interval config not found")
	assert.Equal(t, 654, pollInterval)
	time.Sleep(1 * time.Second)

	pub.Nodes.SetNodeConfigValues(cam1.Address,
		iotc.NodeAttrMap{iotc.NodeAttrPollInterval: "33"},
	)
	time.Sleep(1 * time.Second)
	pollInterval2, _ := pub.GetNodeConfigInt(cam1Id, iotc.NodeAttrPollInterval, 600)
	assert.Equal(t, 33, pollInterval2)

	pub.Stop()
}

// TestStartStop of the ipcam
func TestStartStop(t *testing.T) {
	pub, _ := publisher.NewAppPublisher(AppID, configFolder, cacheFolder, appConfig, false)
	NewIPCamApp(appConfig, pub)

	pub.Start()

	camera := pub.GetNodeByID(cam3Id)
	assert.NotNil(t, camera) // camera node has to exist
	time.Sleep(50 * time.Second)
	pub.Stop()
}
