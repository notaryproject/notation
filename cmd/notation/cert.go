package main

import (
	"errors"
	"fmt"
	"strings"

	corex509 "github.com/notaryproject/notation-core-go/x509"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verification"
	"github.com/notaryproject/notation/internal/truststore"
	"github.com/spf13/cobra"
)

type certAddOpts struct {
	storeType  string
	namedStore string
	path       []string
}

type certListOpts struct {
	storeType  string
	namedStore string
}

type certShowOpts struct {
	storeType  string
	namedStore string
	cert       string
}

type certDeleteOpts struct {
	storeType  string
	namedStore string
	cert       string
	all        bool
	confirmed  bool
}

type certGenerateTestOpts struct {
	name      string
	bits      int
	trust     bool
	isDefault bool
}

func certCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "certificate",
		Aliases: []string{"cert"},
		Short:   "Manage certificates in trust store for signature verification.",
	}

	command.AddCommand(certAddCommand(nil), certListCommand(nil), certShowCommand(nil), certDeleteCommand(nil), certGenerateTestCommand(nil))
	return command
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return addCerts(opts)
		},
	}
	command.Flags().StringVarP(&opts.storeType, "type", "t", "", "specify trust store type, options: ca, signingAuthority")
	command.Flags().StringVarP(&opts.namedStore, "store", "s", "", "specify named store")
	return command
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

func certShowCommand(opts *certShowOpts) *cobra.Command {
	if opts == nil {
		opts = &certShowOpts{}
	}
	command := &cobra.Command{
		Use:   "show --type <type> --store <name> [flags] <cert_fileName>",
		Short: "Show certificate details given trust store type, named store, and certificate file name. If the certificate file contains multiple certificates, then all certificates are displayed.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing certificate path")
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

func certGenerateTestCommand(opts *certGenerateTestOpts) *cobra.Command {
	if opts == nil {
		opts = &certGenerateTestOpts{}
	}
	command := &cobra.Command{
		Use:   "generate-test [flags] <common_name>",
		Short: "Generate a test RSA key and a corresponding self-signed certificate.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing certificate common_name")
			}
			opts.name = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateTestCert(opts)
		},
	}

	command.Flags().IntVarP(&opts.bits, "bits", "b", 2048, "RSA key bits")
	command.Flags().BoolVar(&opts.trust, "trust", false, "add the generated certificate to the trust store")
	setKeyDefaultFlag(command.Flags(), &opts.isDefault)
	return command
}

func addCerts(opts *certAddOpts) error {
	storeType := strings.TrimSpace(opts.storeType)
	if storeType == "" {
		return errors.New("store type cannot be empty or contain only whitespaces")
	}
	if !truststore.ValidateStoreType(storeType) {
		return fmt.Errorf("unsupported store type: %s", storeType)
	}
	namedStore := strings.TrimSpace(opts.namedStore)
	if namedStore == "" {
		return errors.New("named store cannot be empty or contain only whitespaces")
	}
	var success []string
	var failure []string
	var errorSlice []error
	for _, p := range opts.path {
		err := truststore.AddCertCore(p, storeType, namedStore, false)
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
		fmt.Printf("Failed to add following certificates to named store %s of type %s:\n", namedStore, storeType)

		for ind := range failure {
			fmt.Printf("%s, with error %q\n", failure[ind], errorSlice[ind])
		}
	}

	return nil
}

func listCerts(opts *certListOpts) error {
	namedStore := strings.TrimSpace(opts.namedStore)
	storeType := strings.TrimSpace(opts.storeType)

	// List all certificates under truststore/x509, display empty if there's
	// no certificate yet
	if namedStore == "" && storeType == "" {
		path, err := dir.Path.UserConfigFS.GetPath(dir.TrustStoreDir, "x509")
		if err := truststore.CheckError(err); err != nil {
			return err
		}
		if err := truststore.CheckError(truststore.ListCertsCore(path)); err != nil {
			return fmt.Errorf("failed to list all certificates stored in the trust store, with error: %s", err.Error())
		}

		return nil
	}

	// List all certificates under truststore/x509/storeType/namedStore,
	// display empty if there's no such certificate
	if namedStore != "" && storeType != "" {
		path, err := dir.Path.UserConfigFS.GetPath(dir.TrustStoreDir, "x509", storeType, namedStore)
		if err := truststore.CheckError(err); err != nil {
			return err
		}
		if err := truststore.CheckError(truststore.ListCertsCore(path)); err != nil {
			return fmt.Errorf("failed to list certificates stored in the named store %s of type %s, with error: %s", namedStore, storeType, err.Error())
		}

		return nil
	}

	// List all certificates under x509/storeType, display empty if
	// there's no certificate yet
	if storeType != "" {
		path, err := dir.Path.UserConfigFS.GetPath(dir.TrustStoreDir, "x509", storeType)
		if err := truststore.CheckError(err); err != nil {
			return err
		}
		if err := truststore.CheckError(truststore.ListCertsCore(path)); err != nil {
			return fmt.Errorf("failed to list certificates stored of type %s, with error: %s", storeType, err.Error())
		}
	} else {
		// List all certificates under named store namedStore, display empty if
		// there's no such certificate
		for _, t := range verification.TrustStorePrefixes {
			path, err := dir.Path.UserConfigFS.GetPath(dir.TrustStoreDir, "x509", string(t), namedStore)
			if err := truststore.CheckError(err); err != nil {
				return err
			}
			if err := truststore.CheckError(truststore.ListCertsCore(path)); err != nil {
				return fmt.Errorf("failed to list certificates stored in the named store %s, with error: %s", namedStore, err.Error())
			}
		}
	}

	return nil
}

func showCerts(opts *certShowOpts) error {
	storeType := strings.TrimSpace(opts.storeType)
	if storeType == "" {
		return errors.New("store type cannot be empty or contain only whitespaces")
	}
	if !truststore.ValidateStoreType(storeType) {
		return fmt.Errorf("unsupported store type: %s", storeType)
	}
	namedStore := strings.TrimSpace(opts.namedStore)
	if namedStore == "" {
		return errors.New("named store cannot be empty or contain only whitespaces")
	}
	cert := strings.TrimSpace(opts.cert)
	if cert == "" {
		return errors.New("certificate fileName cannot be empty or contain only whitespaces")
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
		return fmt.Errorf("failed to show details of certificate %s, with error: no valid certificate presents", cert)
	}

	//write out
	truststore.ShowCertsCore(certs)

	return nil
}

func deleteCerts(opts *certDeleteOpts) error {
	namedStore := strings.TrimSpace(opts.namedStore)
	if namedStore == "" {
		return errors.New("named store cannot be empty or contain only whitespaces")
	}
	storeType := strings.TrimSpace(opts.storeType)
	if storeType == "" {
		return errors.New("store type cannot be empty or contain only whitespaces")
	}
	if !truststore.ValidateStoreType(storeType) {
		return fmt.Errorf("unsupported store type: %s", storeType)
	}
	var errorSlice []error

	if opts.all {
		// Delete all certificates under storeType/namedStore
		errorSlice = truststore.DeleteAllCerts(storeType, namedStore, opts.confirmed, errorSlice)

		// write out on failure
		if len(errorSlice) > 0 {
			fmt.Println("Failed to clear following named stores:")
			for _, err := range errorSlice {
				fmt.Println(err.Error())
			}
		}

		return nil
	}

	// Delete a certain certificate with path storeType/namedStore/cert
	cert := strings.TrimSpace(opts.cert)
	if cert == "" {
		return errors.New("to delete a specific certificate, certificate fileName cannot be empty or contain only whitespaces")
	}
	errorSlice = truststore.DeleteCert(storeType, namedStore, cert, opts.confirmed, errorSlice)
	// write out on failure
	if len(errorSlice) > 0 {
		fmt.Println("Failed to delete following certificates:")
		for _, err := range errorSlice {
			fmt.Println(err.Error())
		}
	}

	return nil
}
