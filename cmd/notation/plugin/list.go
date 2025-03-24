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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"runtime"
	"syscall"
	"text/tabwriter"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/plugin/proto"
	pluginFramework "github.com/notaryproject/notation-plugin-framework-go/plugin"
	"github.com/spf13/cobra"
)

func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"ls"},
		Short:   "List installed plugins",
		Long: `List installed plugins

Example - List installed Notation plugins:
  notation plugin ls
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listPlugins(cmd)
		},
	}
}

func listPlugins(command *cobra.Command) error {
	mgr := plugin.NewCLIManager(dir.PluginFS())
	pluginNames, err := mgr.List(command.Context())
	if err != nil {
		var errPluginDirWalk plugin.PluginDirectoryWalkError
		if errors.As(err, &errPluginDirWalk) {
			pluginDir, _ := dir.PluginFS().SysPath("")
			return fmt.Errorf("%w.\nPlease ensure that the current user has permission to access the plugin directory: %s", errPluginDirWalk, pluginDir)
		}
		return err
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tDESCRIPTION\tVERSION\tCAPABILITIES\tERROR\t")

	var pl plugin.Plugin
	var resp *proto.GetMetadataResponse
	for _, pluginName := range pluginNames {
		pl, err = mgr.Get(command.Context(), pluginName)
		metaData := &proto.GetMetadataResponse{}
		if err == nil {
			resp, err = pl.GetMetadata(command.Context(), &proto.GetMetadataRequest{})
			if err == nil {
				metaData = resp
			}
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%v\t%v\t\n",
			pluginName, metaData.Description, metaData.Version, metaData.Capabilities, userFriendlyError(pluginName, err))
	}
	return tw.Flush()
}

// userFriendlyError optimizes the error message for the user.
func userFriendlyError(pluginName string, err error) error {
	if err == nil {
		return nil
	}
	var pathError *fs.PathError
	if errors.As(err, &pathError) {
		pluginFileName := pluginFramework.BinaryPrefix + pluginName
		if runtime.GOOS == "windows" {
			pluginFileName += ".exe"
		}

		// for plugin does not exist
		if errors.Is(pathError, fs.ErrNotExist) {
			return fmt.Errorf("%w. Plugin executable file `%s` not found. Use `notation plugin install` command to install the plugin", pathError, pluginFileName)
		}

		// for plugin is not executable
		if pathError.Err == syscall.ENOEXEC {
			return fmt.Errorf("%w. Plugin executable file `%s` is not executable. Use `notation plugin install` command to install the plugin. Please ensure that the plugin executable file is compatible with %s/%s", pathError, pluginFileName, runtime.GOOS, runtime.GOARCH)
		}
	}
	return err
}
