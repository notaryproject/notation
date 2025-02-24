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

import "github.com/spf13/pflag"

// Verfier contains verifier-related flag values.
type Verifier struct {
	// UserMetadata is user metadata flag values for verification
	UserMetadata userMetadata

	// PluginConfig is plugin config flag values for verification
	PluginConfig pluginConfig
}

// ApplyFlags apply flags and their default values for Verifier flags.
func (opts *Verifier) ApplyFlags(fs *pflag.FlagSet) {
	fs.StringArrayVarP((*[]string)(&opts.UserMetadata), userMetadataFlag, "m", nil, "user defined {key}={value} pairs that must be present in the signature for successful verification if provided")
	fs.StringArrayVar((*[]string)(&opts.PluginConfig), pluginConfigFlag, nil, "{key}={value} pairs that are passed as it is to a plugin, if the verification is associated with a verification plugin, refer plugin documentation to set appropriate values")
}
