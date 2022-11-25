package cert

import (
	"errors"
	"fmt"

	"github.com/notaryproject/notation/cmd/notation/internal/truststore"
	"github.com/spf13/cobra"
)

type certAddOpts struct {
	storeType  string
	namedStore string
	path       []string
}

func certAddCommand(opts *certAddOpts) *cobra.Command {
	if opts == nil {
		opts = &certAddOpts{}
	}
	command := &cobra.Command{
		Use:   "add --type <type> --store <name> [flags] <cert_path>...",
		Short: "Add certificates to the trust store.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing certificate path")
			}
			opts.path = args
			return nil
		},
		Long: `Manage certificates in trust store

Example - Add certificates with type "ca" to the trust store "acme-rockets":
  notation cert add --type ca --store "acme-rockets" $XDG_CONFIG_HOME/notation/truststore/x509/ca/acme-rockets/

Example - Add certificates with type "signingAuthority" to the trust store "wabbit-networks":
  notation cert add --type ca --store "wabbit-networks" $XDG_CONFIG_HOME/notation/truststore/x509/signingAuthority/wabbit-networks/
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return addCerts(opts)
		},
	}
	command.Flags().StringVarP(&opts.storeType, "type", "t", "", "specify trust store type, options: ca, signingAuthority")
	command.Flags().StringVarP(&opts.namedStore, "store", "s", "", "specify named store")
	return command
}

func addCerts(opts *certAddOpts) error {
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
	var success []string
	var failure []string
	var errorSlice []error
	for _, p := range opts.path {
		err := truststore.AddCert(p, storeType, namedStore, false)
		if err != nil {
			failure = append(failure, p)
			errorSlice = append(errorSlice, err)
		} else {
			success = append(success, p)
		}
	}

	//write out
	if len(success) != 0 {
		fmt.Printf("Successfully added following certificates to named store %s of type %s:\n", namedStore, storeType)
		for _, p := range success {
			fmt.Println(p)
		}
	}
	if len(failure) != 0 {
		errStr := fmt.Sprintf("Failed to add following certificates to named store %s of type %s:\n", namedStore, storeType)

		for ind := range failure {
			errStr = errStr + fmt.Sprintf("%s, with error %q\n", failure[ind], errorSlice[ind])
		}
		return errors.New(errStr)
	}

	return nil
}
