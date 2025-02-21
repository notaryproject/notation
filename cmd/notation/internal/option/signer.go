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
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/pkg/configutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Signer contains flag options for signing.
type Signer struct {
	Plugin
	Key             string
	SignatureFormat string
}

// ApplyFlags set flags and their default values for the FlagSet.
func (opts *Signer) ApplyFlags(cmd *cobra.Command) {
	opts.Plugin.ApplyFlags(cmd)

	fs := cmd.Flags()
	fs.StringVarP(&opts.Key, "key", "k", "", "signing key name, for a key previously added to notation's key list. This is mutually exclusive with the --id and --plugin flags")
	cmd.MarkFlagsMutuallyExclusive("key", "id")
	cmd.MarkFlagsMutuallyExclusive("key", "plugin")
	opts.setSignatureFormat(fs)
}

func (opts *Signer) setSignatureFormat(fs *pflag.FlagSet) {
	const name = "signature-format"
	const usage = "signature envelope format, options: \"jws\", \"cose\""

	config, err := configutil.LoadConfigOnce()
	if err != nil || config.SignatureFormat == "" {
		// set signatureFormat default to JWS if config is not available
		fs.StringVar(&opts.SignatureFormat, name, envelope.JWS, usage)
		return
	}

	// set signatureFormat from config
	fs.StringVar(&opts.SignatureFormat, name, config.SignatureFormat, usage)
}
