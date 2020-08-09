// Package internal with ipcam to IP based cameras.
// Each camera device is published with an image sensor
package internal

import (
	"strconv"

	"github.com/iotdomain/iotdomain-go/publisher"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// AppID application name used for configuration file and default publisherID
const AppID = "ipcam"

// IPCamConfig with application state, loaded from ipcam.yaml
type IPCamConfig struct {
	PublisherID string `yaml:"publisherId"` // default publisher is app ID
	ImageFolder string `yaml:"imageFolder"` // location to save images
	Cameras     map[string]struct {
		URL          string `yaml:"url"`          // IP camera URL
		PollInterval int    `yaml:"pollInterval"` // how often to poll the camera
		Description  string `yaml:"description"`  // Description of the camera
		Filename     string `yaml:"filename"`     // Locally save image from the camera
		LoginName    string `yaml:"loginName"`
		Password     string `yaml:"password"`
		Overlay      struct {
			Text string `yaml:"text"` // text to overlay
			X    int    `yaml:"x"`    // from left
			Y    int    `yaml:"y"`    // from bottom
			Size int    `yaml:"size"` // font size in pt
		} `yaml:"overlay"`
	} `yaml:"cameras"`
}

// IPCamApp publisher app
type IPCamApp struct {
	config    *IPCamConfig
	pub       *publisher.Publisher
	pollDelay map[string]int // seconds until next poll for each camera
}

// CreateCamerasFromConfig loads cameras from config and add outputs for image and latency.
func (ipcam *IPCamApp) CreateCamerasFromConfig(config *IPCamConfig) {
	pub := ipcam.pub
	logrus.Infof("Loading %d cameras from config", len(config.Cameras))

	for camID, camInfo := range config.Cameras {
		// node := pub.GetNodeByID(camID)
		// if node == nil {
		pub.CreateNode(camID, types.NodeTypeCamera)
		pub.UpdateNodeAttr(camID, types.NodeAttrMap{
			types.NodeAttrDescription: camInfo.Description,
			types.NodeAttrFilename:    camInfo.Filename,
		})

		pub.UpdateNodeConfig(camID, types.NodeAttrURL, &types.ConfigAttr{
			DataType:    types.DataTypeString,
			Description: "Camera URL, for example http://images.drivebc.ca/bchighwaycam/pub/cameras/2.jpg",
			Default:     camInfo.URL,
		})
		pub.UpdateNodeConfig(camID, types.NodeAttrLoginName, &types.ConfigAttr{
			DataType:    types.DataTypeString,
			Description: "Camera login name",
			Default:     camInfo.LoginName,
			Secret:      true, // don't include value in discovery publication
		})
		pub.UpdateNodeConfig(camID, types.NodeAttrPassword, &types.ConfigAttr{
			DataType:    types.DataTypeString,
			Description: "Camera password",
			Default:     camInfo.Password,
			Secret:      true, // don't include value in discovery publication
		})
		// each camera has its own poll interval
		pub.UpdateNodeConfig(camID, types.NodeAttrPollInterval, &types.ConfigAttr{
			DataType:    types.DataTypeInt,
			Description: "Camera poll interval in seconds",
			Default:     strconv.Itoa(camInfo.PollInterval),
			Min:         5,
			Max:         3600,
		})
		// the image and camera latency are both outputs
		pub.CreateOutput(camID, types.OutputTypeImage, types.DefaultOutputInstance)
		pub.CreateOutput(camID, types.OutputTypeLatency, types.DefaultOutputInstance)

		pub.CreateInputFromHTTP(camID, types.InputTypeImage, types.DefaultInputInstance,
			camInfo.URL, camInfo.LoginName, camInfo.Password, camInfo.PollInterval, ipcam.handleInputMessage)
	}

}

// NewIPCamApp creates the app
func NewIPCamApp(config *IPCamConfig, pub *publisher.Publisher) *IPCamApp {
	app := IPCamApp{
		config:    config,
		pub:       pub,
		pollDelay: make(map[string]int),
	}
	if app.config.PublisherID == "" {
		app.config.PublisherID = AppID
	}
	app.CreateCamerasFromConfig(config)

	pub.SetPollInterval(1, app.Poll)
	// pub.SetNodeInputHandler(app.HandleInputCommand)
	pub.SetNodeConfigHandler(app.HandleConfigCommand)
	// // Discover the node(s) and outputs. Use default for republishing discovery
	// onewirePub.SetDiscoveryInterval(0, app.Discover)
	return &app
}

// Run the publisher until the SIGTERM  or SIGINT signal is received
func Run() {
	appConfig := &IPCamConfig{PublisherID: AppID}
	pub, _ := publisher.NewAppPublisher(AppID, "", appConfig, true)

	app := NewIPCamApp(appConfig, pub)
	_ = app

	pub.Start()
	pub.WaitForSignal()
	pub.Stop()
}
