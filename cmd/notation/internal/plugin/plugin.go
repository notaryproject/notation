package plugin

import (
	"context"
	"errors"
	"os"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
)

// CheckPluginExistence returns true if a plugin already exists
func CheckPluginExistence(ctx context.Context, pluginName string) (bool, error) {
	mgr := plugin.NewCLIManager(dir.PluginFS())
	_, err := mgr.Get(ctx, pluginName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
