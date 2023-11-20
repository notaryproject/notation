package plugin

import (
	"context"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/plugin/proto"
)

// GetPluginMetadataIfExist returns plugin's metadata if it exists in Notation
func GetPluginMetadataIfExist(ctx context.Context, pluginName string) (*proto.GetMetadataResponse, error) {
	mgr := plugin.NewCLIManager(dir.PluginFS())
	plugin, err := mgr.Get(ctx, pluginName)
	if err != nil {
		return nil, err
	}
	return plugin.GetMetadata(ctx, &proto.GetMetadataRequest{})
}
