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
	"os"

	"github.com/spf13/pflag"
	"oras.land/oras-go/v2/registry/remote/auth"
)

const (
	DefaultUsernameEnv = "NOTATION_USERNAME"
	DefaultPasswordEnv = "NOTATION_PASSWORD"
	defaultMediaType   = "application/vnd.docker.distribution.manifest.v2+json"
)

// Secure contains flag options for registry authentication.
type Secure struct {
	// Username for registry authentication.
	Username string

	// Password for registry authentication.
	Password string

	// InsecureRegistry indicates whether to skip TLS verification.
	InsecureRegistry bool
}

// ApplyFlags set flags and their default values for the FlagSet
func (opts *Secure) ApplyFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&opts.Username, "username", "u", "", "username for registry operations (default to $NOTATION_USERNAME if not specified)")
	fs.StringVarP(&opts.Password, "password", "p", "", "password for registry operations (default to $NOTATION_PASSWORD if not specified)")
	fs.BoolVar(&opts.InsecureRegistry, "insecure-registry", false, "use HTTP protocol while connecting to registries. Should be used only for testing")
	opts.Username = os.Getenv(DefaultUsernameEnv)
	opts.Password = os.Getenv(DefaultPasswordEnv)
}

// Credential returns an auth.Credential from opts.Username and opts.Password.
func (opts *Secure) Credential() auth.Credential {
	if opts.Username == "" {
		return auth.Credential{
			RefreshToken: opts.Password,
		}
	}
	return auth.Credential{
		Username: opts.Username,
		Password: opts.Password,
	}
}
