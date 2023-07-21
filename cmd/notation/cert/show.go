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

package cert

import (
	"context"
	"errors"
	"fmt"

	corex509 "github.com/notaryproject/notation-core-go/x509"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/log"
	"github.com/notaryproject/notation/cmd/notation/internal/truststore"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/spf13/cobra"
)

type certShowOpts struct {
	cmd.LoggingFlagOpts
	storeType  string
	namedStore string
	cert       string
}

func certShowCommand(opts *certShowOpts) *cobra.Command {
	if opts == nil {
		opts = &certShowOpts{}
	}
	command := &cobra.Command{
		Use:   "show --type <type> --store <name> [flags] <cert_fileName>",
		Short: "Show certificate details given trust store type, named store, and certificate file name. If the certificate file contains multiple certificates, then all certificates are displayed.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing certificate file name")
			}
			if len(args) > 1 {
				return errors.New("show only supports single certificate file")
			}
			opts.cert = args[0]
			return nil
		},
		Long: `Show certificate details of given trust store name, trust store type, and certificate file name. If the certificate file contains multiple certificates, then all certificates are displayed

Example - Show details of certificate "cert1.pem" with type "ca" from trust store "acme-rockets":
  notation cert show --type ca --store acme-rockets cert1.pem

Example - Show details of certificate "cert2.pem" with type "signingAuthority" from trust store "wabbit-networks":
  notation cert show --type signingAuthority --store wabbit-networks cert2.pem
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCerts(cmd.Context(), opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	command.Flags().StringVarP(&opts.storeType, "type", "t", "", "specify trust store type, options: ca, signingAuthority")
	command.Flags().StringVarP(&opts.namedStore, "store", "s", "", "specify named store")
	command.MarkFlagRequired("type")
	command.MarkFlagRequired("store")
	return command
}

func showCerts(ctx context.Context, opts *certShowOpts) error {
	// set log level
	ctx = opts.LoggingFlagOpts.InitializeLogger(ctx)
	logger := log.GetLogger(ctx)

	storeType := opts.storeType
	if storeType == "" {
		return errors.New("store type cannot be empty")
	}
	if !truststore.IsValidStoreType(storeType) {
		return fmt.Errorf("unsupported store type: %s", storeType)
	}
	namedStore := opts.namedStore
	if !truststore.IsValidFileName(namedStore) {
		return errors.New("named store name needs to follow [a-zA-Z0-9_.-]+ format")
	}
	cert := opts.cert
	if cert == "" {
		return errors.New("certificate fileName cannot be empty")
	}

	path, err := dir.ConfigFS().SysPath(dir.TrustStoreDir, "x509", storeType, namedStore, cert)
	if err != nil {
		return fmt.Errorf("failed to show details of certificate %s, with error: %s", cert, err.Error())
	}
	logger.Debugln("Showing details of certificate:", path)
	certs, err := corex509.ReadCertificateFile(path)
	if err != nil {
		return fmt.Errorf("failed to show details of certificate %s, with error: %s", cert, err.Error())
	}
	if len(certs) == 0 {
		return fmt.Errorf("failed to show details of certificate %s, with error: no valid certificate found in the file", cert)
	}

	//write out
	truststore.ShowCerts(certs)

	return nil
}
