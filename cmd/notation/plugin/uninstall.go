// Copyright The Notary Project Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugin

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation/cmd/notation/internal/cmdutil"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/spf13/cobra"
)

type pluginUninstallOpts struct {
	cmd.LoggingFlagOpts
	pluginName string
	confirmed  bool
}

func uninstallCommand(opts *pluginUninstallOpts) *cobra.Command {
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
				return errors.New("only one plugin can be removed at a time")
			}
			opts.pluginName = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return uninstallPlugin(cmd, opts)
		},
	}

	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	command.Flags().BoolVarP(&opts.confirmed, "yes", "y", false, "do not prompt for confirmation")
	return command
}

func uninstallPlugin(command *cobra.Command, opts *pluginUninstallOpts) error {
	// set logger
	ctx := opts.LoggingFlagOpts.InitializeLogger(command.Context())
	pluginName := opts.pluginName
	exist, err := checkPluginExistence(ctx, pluginName)
	if err != nil {
		return fmt.Errorf("failed to check plugin existence: %w", err)
	}
	if !exist {
		return fmt.Errorf("unable to find plugin %s.\nTo view a list of installed plugins, use `notation plugin list`", pluginName)
	}
	// core process
	prompt := fmt.Sprintf("Are you sure you want to uninstall plugin %q?", pluginName)
	confirmed, err := cmdutil.AskForConfirmation(os.Stdin, prompt, opts.confirmed)
	if err != nil {
		return fmt.Errorf("failed when asking for confirmation: %w", err)
	}
	if !confirmed {
		return nil
	}
	mgr := plugin.NewCLIManager(dir.PluginFS())
	if err := mgr.Uninstall(ctx, pluginName); err != nil {
		return fmt.Errorf("failed to uninstall plugin %s: %w", pluginName, err)
	}
	fmt.Printf("Successfully uninstalled plugin %s\n", pluginName)
	return nil
}

// checkPluginExistence returns true if plugin exists in the system
func checkPluginExistence(ctx context.Context, pluginName string) (bool, error) {
	mgr := plugin.NewCLIManager(dir.PluginFS())
	_, err := mgr.Get(ctx, pluginName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) { // plugin does not exist
			return false, nil
		}
		return false, err
	}
	return true, nil
}
