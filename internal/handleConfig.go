// Package internal handles node configuration commands
package internal

import "github.com/iotdomain/iotdomain-go/types"

// HandleConfigCommand handles requests to update node configuration
func (app *IPCamApp) HandleConfigCommand(node *types.NodeDiscoveryMessage, config types.NodeAttrMap) types.NodeAttrMap {
	app.logger.Infof("IPCamApp.HandleConfigCommand for %s. Ignored as this isn't yet supported", node.Address)
	return nil
}
