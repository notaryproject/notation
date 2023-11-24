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
	"github.com/notaryproject/notation-go/log"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/plugin/proto"
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
			return unInstallPlugin(cmd, opts)
		},
	}

	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	command.Flags().BoolVarP(&opts.confirmed, "yes", "y", false, "do not prompt for confirmation")
	return command
}

func unInstallPlugin(command *cobra.Command, opts *pluginUninstallOpts) error {
	// set logger
	ctx := opts.LoggingFlagOpts.InitializeLogger(command.Context())
	logger := log.GetLogger(ctx)

	pluginName := opts.pluginName
	_, err := getPluginMetadataIfExist(ctx, pluginName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) { // plugin does not exist
			return fmt.Errorf("unable to find plugin %s.\nTo view a list of installed plugins, use `notation plugin list`", pluginName)
		}
		// plugin exists, but the binary is malfunctioning
		logger.Infof("Uninstalling...Found plugin %s binary file is malfunctioning: %v", pluginName, err)
	}
	// core process
	prompt := fmt.Sprintf("Are you sure you want to uninstall plugin %q?", pluginName)
	confirmed, err := cmdutil.AskForConfirmation(os.Stdin, prompt, opts.confirmed)
	if err != nil {
		return fmt.Errorf("failed when asking for confirmation: %v", err)
	}
	if !confirmed {
		return nil
	}
	mgr := plugin.NewCLIManager(dir.PluginFS())
	if err := mgr.Uninstall(ctx, pluginName); err != nil {
		return fmt.Errorf("failed to uninstall %s: %v", pluginName, err)
	}
	fmt.Printf("Successfully uninstalled plugin %s\n", pluginName)
	return nil
}

// getPluginMetadataIfExist returns plugin's metadata if it exists in Notation
func getPluginMetadataIfExist(ctx context.Context, pluginName string) (*proto.GetMetadataResponse, error) {
	mgr := plugin.NewCLIManager(dir.PluginFS())
	plugin, err := mgr.Get(ctx, pluginName)
	if err != nil {
		return nil, err
	}
	return plugin.GetMetadata(ctx, &proto.GetMetadataRequest{})
}
