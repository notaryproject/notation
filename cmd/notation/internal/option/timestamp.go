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
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Timestamp contains timestamp-related flag values
type Timestamp struct {
	// ServerURL is the URL of the Timestamping Authority (TSA) server.
	ServerURL string

	// RootCertificatePath is the file path of the TSA root certificate.
	RootCertificatePath string
}

// ApplyFlags apply flags and their default values for Timestamp flags.
func (opts *Timestamp) ApplyFlags(fs *pflag.FlagSet) {
	fs.StringVar(&opts.ServerURL, "timestamp-url", "", "RFC 3161 Timestamping Authority (TSA) server URL")
	fs.StringVar(&opts.RootCertificatePath, "timestamp-root-cert", "", "filepath of timestamp authority root certificate")
}

// Validate validates Timestamp flags.
func (opts *Timestamp) Validate(cmd *cobra.Command) error {
	if cmd.Flags().Changed("timestamp-url") {
		if opts.ServerURL == "" {
			return errors.New("timestamping: tsa url cannot be empty")
		}
		if opts.RootCertificatePath == "" {
			return errors.New("timestamping: tsa root certificate path cannot be empty")
		}
	}
	return nil
}
