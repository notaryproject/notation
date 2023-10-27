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
	"fmt"
	"os"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/log"
	notationgoTruststore "github.com/notaryproject/notation-go/verifier/truststore"
	"github.com/notaryproject/notation/cmd/notation/internal/truststore"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/spf13/cobra"
)

type certListOpts struct {
	cmd.LoggingFlagOpts
	storeType  string
	namedStore string
}

func certListCommand(opts *certListOpts) *cobra.Command {
	if opts == nil {
		opts = &certListOpts{}
	}
	command := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List certificates in the trust store.",
		Long: `List certificates in the trust store

Example - List all certificate files stored in the trust store
  notation cert ls

Example - List all certificate files of trust store "acme-rockets"
  notation cert ls --store "acme-rockets"

Example - List all certificate files from trust store of type "ca"
  notation cert ls --type ca

Example - List all certificate files from trust store "wabbit-networks" of type "signingAuthority"
  notation cert ls --type signingAuthority --store "wabbit-networks"
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listCerts(cmd.Context(), opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	command.Flags().StringVarP(&opts.storeType, "type", "t", "", "specify trust store type, options: ca, signingAuthority")
	command.Flags().StringVarP(&opts.namedStore, "store", "s", "", "specify named store")
	return command
}

func listCerts(ctx context.Context, opts *certListOpts) error {
	// set log level
	ctx = opts.LoggingFlagOpts.InitializeLogger(ctx)
	logger := log.GetLogger(ctx)

	namedStore := opts.namedStore
	storeType := opts.storeType
	configFS := dir.ConfigFS()

	// List all certificates under truststore/x509, display empty if there's
	// no certificate yet
	if namedStore == "" && storeType == "" {
		var certPaths []string
		for _, t := range notationgoTruststore.Types {
			path, err := configFS.SysPath(dir.TrustStoreDir, "x509", string(t))
			if err := truststore.CheckNonErrNotExistError(err); err != nil {
				return err
			}
			certs, err := truststore.ListCerts(path, 1)
			if err := truststore.CheckNonErrNotExistError(err); err != nil {
				logger.Debugln("Failed to complete list at path:", path)
				return fmt.Errorf("failed to list all certificates stored in the trust store, with error: %s", err.Error())
			}
			certPaths = append(certPaths, certs...)
		}
		return ioutil.PrintCertMap(os.Stdout, certPaths)
	}

	// List all certificates under truststore/x509/storeType/namedStore,
	// display empty if store type is invalid or there's no certificate yet
	if namedStore != "" && storeType != "" {
		if !truststore.IsValidStoreType(storeType) {
			return nil
		}
		path, err := configFS.SysPath(dir.TrustStoreDir, "x509", storeType, namedStore)
		if err := truststore.CheckNonErrNotExistError(err); err != nil {
			return err
		}
		certPaths, err := truststore.ListCerts(path, 0)
		if err := truststore.CheckNonErrNotExistError(err); err != nil {
			logger.Debugln("Failed to complete list at path:", path)
			return fmt.Errorf("failed to list all certificates stored in the named store %s of type %s, with error: %s", namedStore, storeType, err.Error())
		}
		return ioutil.PrintCertMap(os.Stdout, certPaths)
	}

	// List all certificates under x509/storeType, display empty if store type
	// is invalid or there's no certificate yet
	if storeType != "" {
		if !truststore.IsValidStoreType(storeType) {
			return nil
		}
		path, err := configFS.SysPath(dir.TrustStoreDir, "x509", storeType)
		if err := truststore.CheckNonErrNotExistError(err); err != nil {
			return err
		}
		certPaths, err := truststore.ListCerts(path, 1)
		if err := truststore.CheckNonErrNotExistError(err); err != nil {
			logger.Debugln("Failed to complete list at path:", path)
			return fmt.Errorf("failed to list all certificates stored of type %s, with error: %s", storeType, err.Error())
		}
		return ioutil.PrintCertMap(os.Stdout, certPaths)
	}

	// List all certificates under named store namedStore, display empty if
	// there's no certificate yet
	var certPaths []string
	for _, t := range notationgoTruststore.Types {
		path, err := configFS.SysPath(dir.TrustStoreDir, "x509", string(t), namedStore)
		if err := truststore.CheckNonErrNotExistError(err); err != nil {
			return err
		}
		certs, err := truststore.ListCerts(path, 0)
		if err := truststore.CheckNonErrNotExistError(err); err != nil {
			logger.Debugln("Failed to complete list at path:", path)
			return fmt.Errorf("failed to list all certificates stored in the named store %s, with error: %s", namedStore, err.Error())
		}
		certPaths = append(certPaths, certs...)
	}
	return ioutil.PrintCertMap(os.Stdout, certPaths)
}
