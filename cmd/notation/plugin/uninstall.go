package plugin

import (
	"errors"
	"fmt"
	"os"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation/cmd/notation/internal/cmdutil"
	notationplugin "github.com/notaryproject/notation/cmd/notation/internal/plugin"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/spf13/cobra"
)

type pluginUninstallOpts struct {
	cmd.LoggingFlagOpts
	pluginName string
	confirmed  bool
}

func pluginUninstallCommand(opts *pluginUninstallOpts) *cobra.Command {
	if opts == nil {
		opts = &pluginUninstallOpts{}
	}
	command := &cobra.Command{
		Use:     "uninstall [flags] <plugin_name>",
		Aliases: []string{"remove", "rm"},
		Short:   "Uninstall a plugin",
		Long: `Uninstall a plugin

Example - Uninstall plugin:
  notation plugin uninstall wabbit-plugin
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("plugin name is required")
			}
			if len(args) > 1 {
				return errors.New("can only remove one plugin at a time")
			}
			opts.pluginName = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return unInstallPlugin(cmd, opts)
		},
	}

	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	command.Flags().BoolVarP(&opts.confirmed, "yes", "y", false, "do not prompt for confirmation")
	return command
}

func unInstallPlugin(command *cobra.Command, opts *pluginUninstallOpts) error {
	// set log level
	ctx := opts.LoggingFlagOpts.InitializeLogger(command.Context())
	pluginName := opts.pluginName
	_, err := notationplugin.GetPluginMetadataIfExist(ctx, pluginName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) { // plugin does not exist
			return fmt.Errorf("unable to find plugin %s.\nTo view a list of installed plugins, use `notation plugin list`", pluginName)
		}
		return fmt.Errorf("failed to uninstall %s: %w", pluginName, err)
	}
	// core process
	pluginPath, err := dir.PluginFS().SysPath(pluginName)
	if err != nil {
		return fmt.Errorf("failed to uninstall %s: %v", pluginName, err)
	}
	prompt := fmt.Sprintf("Are you sure you want to uninstall plugin %q?", pluginName)
	confirmed, err := cmdutil.AskForConfirmation(os.Stdin, prompt, opts.confirmed)
	if err != nil {
		return fmt.Errorf("failed to uninstall %s: %v", pluginName, err)
	}
	if !confirmed {
		return nil
	}
	if err := os.RemoveAll(pluginPath); err != nil {
		return fmt.Errorf("failed to uninstall %s: %v", pluginName, err)
	}
	fmt.Printf("Successfully uninstalled plugin %s\n", pluginName)
	return nil
}
