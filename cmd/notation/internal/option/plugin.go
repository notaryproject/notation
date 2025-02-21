package option

import (
	"github.com/spf13/cobra"
)

// Plugin contains key-related flag values
type Plugin struct {
	PluginConfig
	PluginName string
	KeyID      string
}

// ApplyFlags set flags and their default values for the FlagSet.
func (opts *Plugin) ApplyFlags(cmd *cobra.Command) {
	fs := cmd.Flags()
	opts.PluginConfig.ApplyFlags(fs)
	fs.StringVar(&opts.KeyID, "id", "", "key id (required if --plugin is set). This is mutually exclusive with the --key flag")
	fs.StringVar(&opts.PluginName, "plugin", "", "signing plugin name (required if --id is set). This is mutually exclusive with the --key flag")
	cmd.MarkFlagsRequiredTogether("id", "plugin")
}
