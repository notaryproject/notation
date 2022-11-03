package cert

import (
	"errors"
	"fmt"

	corex509 "github.com/notaryproject/notation-core-go/x509"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation/cmd/notation/internal/truststore"
	"github.com/spf13/cobra"
)

type certShowOpts struct {
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCerts(opts)
		},
	}
	command.Flags().StringVarP(&opts.storeType, "type", "t", "", "specify trust store type, options: ca, signingAuthority")
	command.Flags().StringVarP(&opts.namedStore, "store", "s", "", "specify named store")
	return command
}

func showCerts(opts *certShowOpts) error {
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

	path, err := dir.Path.UserConfigFS.GetPath(dir.TrustStoreDir, "x509", storeType, namedStore, cert)
	if err != nil {
		return fmt.Errorf("failed to show details of certificate %s, with error: %s", cert, err.Error())
	}
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
