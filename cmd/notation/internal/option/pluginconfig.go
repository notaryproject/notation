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
	"github.com/spf13/pflag"
)

const pluginConfigFlag = "plugin-config"

// PluginConfig is a plugin config option.
type PluginConfig []string

// ApplyFlags sets up the flags for the plugin config option.
func (pc *PluginConfig) ApplyFlags(fs *pflag.FlagSet) {
	fs.StringArrayVar((*[]string)(pc), pluginConfigFlag, nil, "{key}={value} pairs that are passed as it is to a plugin, refer plugin's documentation to set appropriate values")
}

// ParseFlagMap parses plugin-config flag into a map.
func (pc *PluginConfig) PluginConfigMap() (map[string]string, error) {
	return parseFlagMap(*pc, pluginConfigFlag)
}

// VerificationPluginConfig contains a plugin config for verification.
type VerificationPluginConfig struct {
	PluginConfig
}

// ApplyFlags sets up the flags for the verification plugin config option.
func (vpc *VerificationPluginConfig) ApplyFlags(fs *pflag.FlagSet) {
	fs.StringArrayVar((*[]string)(&vpc.PluginConfig), pluginConfigFlag, nil, "{key}={value} pairs that are passed as it is to a plugin, if the verification is associated with a verification plugin, refer plugin documentation to set appropriate values")
}
