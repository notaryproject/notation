package cert

import (
	"errors"
	"fmt"
	"strings"

	"github.com/notaryproject/notation/cmd/notation/internal/truststore"
	"github.com/spf13/cobra"
)

type Errors []error

func (errs Errors) Error() string {
	var messages []string
	for _, err := range errs {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "\n")
}

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
		Use:   "delete --type <type> --store <name> [flags] (--all | <cert_fileName>)",
		Short: "Delete certificates from the trust store.",
		Args: func(cmd *cobra.Command, args []string) error {
			if !opts.all && len(args) == 0 {
				return errors.New("delete requires either the certificate file name that needs to be deleted or --all flag to delete all certificates in the given named trust store")
			}
			if len(args) != 0 {
				opts.cert = args[0]
			}
			return nil
		},
		Long: `Delete certificates from the trust store

Example - Delete all certificates with "ca" type from the trust store "acme-rockets":
  notation cert delete --type ca --store acme-rockets --all

Example - Delete certificate "cert1.pem" with "signingAuthority" type from trust store wabbit-networks:
  notation cert delete --type signingAuthority --store wabbit-networks cert1.pem

Example - Delete all certificates with "ca" type from the trust store "acme-rockets", without prompt for confirmation:
  notation cert delete --type ca --store acme-rockets -y --all 
`,
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
	if !truststore.IsValidFileName(namedStore) {
		return errors.New("named store name needs to follow [a-zA-Z0-9_.-]+ format")
	}
	storeType := opts.storeType
	if storeType == "" {
		return errors.New("store type cannot be empty")
	}
	if !truststore.IsValidStoreType(storeType) {
		return fmt.Errorf("unsupported store type: %s", storeType)
	}
	var errs Errors

	if opts.all {
		// Delete all certificates under storeType/namedStore
		errs = truststore.DeleteAllCerts(storeType, namedStore, opts.confirmed, errs)
		// write out on failure
		if len(errs) > 0 {
			errStr := "Failed to delete following named stores:\n"
			errs = append([]error{errors.New(errStr)}, errs...)
			return errs
		}

		return nil
	}

	// Delete a certain certificate with path storeType/namedStore/cert
	cert := opts.cert
	if cert == "" {
		return errors.New("to delete a specific certificate, certificate fileName cannot be empty")
	}
	errs = truststore.DeleteCert(storeType, namedStore, cert, opts.confirmed, errs)
	// write out on failure
	if len(errs) > 0 {
		errStr := "Failed to delete following certificate:\n"
		errs = append([]error{errors.New(errStr)}, errs...)
		return errs
	}

	return nil
}
