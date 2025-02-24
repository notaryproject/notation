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
