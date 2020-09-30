// Package internal handles node configuration commands
package internal

import (
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// HandleConfigCommand handles requests to update node configuration
func (app *IPCamApp) HandleConfigCommand(nodeHWID string, config types.NodeAttrMap) {
	logrus.Infof("IPCamApp.HandleConfigCommand for node %s. Accepting config.", nodeHWID)
	app.pub.UpdateNodeConfigValues(nodeHWID, config)
}
