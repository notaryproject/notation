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

// TrustStore defines a trust store option.
type TrustStore struct {
	StoreType  string
	NamedStore string
}

// ApplyFlags applies store flags with default values for trust store flags.
func (opts *TrustStore) ApplyFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&opts.StoreType, "type", "t", "", "specify trust store type, options: ca, signingAuthority")
	fs.StringVarP(&opts.NamedStore, "store", "s", "", "specify named store")
}
