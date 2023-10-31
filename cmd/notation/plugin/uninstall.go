package plugin

import (
	"errors"
	"fmt"
	"os"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation/cmd/notation/internal/cmdutil"
	notationplugin "github.com/notaryproject/notation/cmd/notation/internal/plugin"
	"github.com/spf13/cobra"
)

type pluginUninstallOpts struct {
	pluginName string
	confirmed  bool
}

func pluginUninstallCommand(opts *pluginUninstallOpts) *cobra.Command {
	if opts == nil {
		opts = &pluginUninstallOpts{}
	}
	command := &cobra.Command{
		Use:   "uninstall [flags] <plugin_name>",
		Short: "Uninstall plugin",
		Long: `Uninstall a Notation plugin

Example - Uninstall plugin:
  notation plugin uninstall my-plugin
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("plugin name is required")
			}
			opts.pluginName = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return unInstallPlugin(cmd, opts)
		},
	}

	command.Flags().BoolVarP(&opts.confirmed, "yes", "y", false, "do not prompt for confirmation")
	return command
}

func unInstallPlugin(command *cobra.Command, opts *pluginUninstallOpts) error {
	pluginName := opts.pluginName
	existed, err := notationplugin.CheckPluginExistence(command.Context(), pluginName)
	if err != nil {
		return fmt.Errorf("failed to check plugin existence, %w", err)
	}
	if !existed {
		return fmt.Errorf("plugin %s does not exist", pluginName)
	}
	pluginPath, err := dir.PluginFS().SysPath(pluginName)
	if err != nil {
		return err
	}
	prompt := fmt.Sprintf("Are you sure you want to uninstall plugin %q?", pluginName)
	confirmed, err := cmdutil.AskForConfirmation(os.Stdin, prompt, opts.confirmed)
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}
	err = os.RemoveAll(pluginPath)
	if err == nil {
		fmt.Printf("Successfully uninstalled plugin %s\n", pluginName)
	}
	return err
}
