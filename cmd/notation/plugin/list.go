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
	"os"
	"text/tabwriter"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/plugin/proto"
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
	for _, n := range pluginNames {
		pl, err = mgr.Get(command.Context(), n)
		metaData := &proto.GetMetadataResponse{}
		if err == nil {
			resp, err = pl.GetMetadata(command.Context(), &proto.GetMetadataRequest{})
			if err == nil {
				metaData = resp
			}
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%v\t%v\t\n",
			n, metaData.Description, metaData.Version, metaData.Capabilities, err)
	}
	return tw.Flush()
}
