// Package internal handles node configuration commands
package internal

import (
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// HandleConfigCommand handles requests to update node configuration
func (app *IPCamApp) HandleConfigCommand(address string, config types.NodeAttrMap) types.NodeAttrMap {
	logrus.Infof("IPCamApp.HandleConfigCommand for %s. Accepting config.", address)
	return config
}
