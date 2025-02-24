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

// IsDefaultKey defines a default option to mark a key as default.
type IsDefaultKey struct {
	IsDefault bool
}

// ApplyFlags applies default flags with default values.
func (opts *IsDefaultKey) ApplyFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&opts.IsDefault, "default", false, "mark as default signing key")
}
