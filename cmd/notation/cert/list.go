package cert

import (
	"fmt"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verification"
	"github.com/notaryproject/notation/cmd/notation/internal/truststore"
	"github.com/spf13/cobra"
)

type certListOpts struct {
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return listCerts(opts)
		},
	}
	command.Flags().StringVarP(&opts.storeType, "type", "t", "", "specify trust store type, options: ca, signingAuthority")
	command.Flags().StringVarP(&opts.namedStore, "store", "s", "", "specify named store")
	return command
}

func listCerts(opts *certListOpts) error {
	namedStore := opts.namedStore
	storeType := opts.storeType

	// List all certificates under truststore/x509, display empty if there's
	// no certificate yet
	if namedStore == "" && storeType == "" {
		path := dir.X509TrustStoreDir()
		if err := truststore.CheckNonErrNotExistError(truststore.ListCerts(path, 2)); err != nil {
			return fmt.Errorf("failed to list all certificates stored in the trust store, with error: %s", err.Error())
		}

		return nil
	}

	// List all certificates under truststore/x509/storeType/namedStore,
	// display empty if there's no such certificate
	if namedStore != "" && storeType != "" {
		path := dir.X509TrustStoreDir(storeType, namedStore)
		if err := truststore.CheckNonErrNotExistError(truststore.ListCerts(path, 0)); err != nil {
			return fmt.Errorf("failed to list certificates stored in the named store %s of type %s, with error: %s", namedStore, storeType, err.Error())
		}

		return nil
	}

	// List all certificates under x509/storeType, display empty if
	// there's no certificate yet
	if storeType != "" {
		path := dir.X509TrustStoreDir(storeType)
		if err := truststore.CheckNonErrNotExistError(truststore.ListCerts(path, 1)); err != nil {
			return fmt.Errorf("failed to list certificates stored of type %s, with error: %s", storeType, err.Error())
		}
	} else {
		// List all certificates under named store namedStore, display empty if
		// there's no such certificate
		for _, t := range verification.TrustStorePrefixes {
			path := dir.X509TrustStoreDir(string(t), namedStore)
			if err := truststore.CheckNonErrNotExistError(truststore.ListCerts(path, 0)); err != nil {
				return fmt.Errorf("failed to list certificates stored in the named store %s, with error: %s", namedStore, err.Error())
			}
		}
	}

	return nil
}
