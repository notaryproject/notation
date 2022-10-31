package cert

import (
	"errors"
	"fmt"

	"github.com/notaryproject/notation/cmd/notation/internal/truststore"
	"github.com/spf13/cobra"
)

type certDeleteOpts struct {
	storeType  string
	namedStore string
	cert       string
	all        bool
	confirmed  bool
}

func certDeleteCommand(opts *certDeleteOpts) *cobra.Command {
	if opts == nil {
		opts = &certDeleteOpts{}
	}
	command := &cobra.Command{
		Use:     "delete --type <type> --store <name> [flags] (--all | <cert_fileName>)",
		Aliases: []string{"rm"},
		Short:   "Delete certificates from the trust store.",
		Args: func(cmd *cobra.Command, args []string) error {
			if !opts.all && len(args) == 0 {
				return errors.New("delete requires either the certificate file name that needs to be deleted or --all flag to delete all certificates in the given named trust store")
			}
			if len(args) != 0 {
				opts.cert = args[0]
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteCerts(opts)
		},
	}
	command.Flags().StringVarP(&opts.storeType, "type", "t", "", "specify trust store type, options: ca, signingAuthority")
	command.Flags().StringVarP(&opts.namedStore, "store", "s", "", "specify named store")
	command.Flags().BoolVarP(&opts.all, "all", "a", false, "delete all certificates in the named store")
	command.Flags().BoolVarP(&opts.confirmed, "yes", "y", false, "do not prompt for confirmation")
	return command
}

func deleteCerts(opts *certDeleteOpts) error {
	namedStore := opts.namedStore
	if namedStore == "" {
		return errors.New("named store cannot be empty")
	}
	storeType := opts.storeType
	if storeType == "" {
		return errors.New("store type cannot be empty")
	}
	if !truststore.IsValidStoreType(storeType) {
		return fmt.Errorf("unsupported store type: %s", storeType)
	}
	var errorSlice []error

	if opts.all {
		// Delete all certificates under storeType/namedStore
		errorSlice = truststore.DeleteAllCerts(storeType, namedStore, opts.confirmed, errorSlice)

		// write out on failure
		if len(errorSlice) > 0 {
			errStr := "Failed to delete following named stores:\n"
			for _, err := range errorSlice {
				errStr = errStr + err.Error() + "\n"
			}
			return errors.New(errStr)
		}

		return nil
	}

	// Delete a certain certificate with path storeType/namedStore/cert
	cert := opts.cert
	if cert == "" {
		return errors.New("to delete a specific certificate, certificate fileName cannot be empty")
	}
	errorSlice = truststore.DeleteCert(storeType, namedStore, cert, opts.confirmed, errorSlice)
	// write out on failure
	if len(errorSlice) > 0 {
		errStr := "Failed to delete following certificate:\n"
		for _, err := range errorSlice {
			errStr = errStr + err.Error() + "\n"
		}
		return errors.New(errStr)
	}

	return nil
}
