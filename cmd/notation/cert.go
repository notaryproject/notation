package main

import (
	"crypto/x509"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	corex509 "github.com/notaryproject/notation-core-go/x509"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/internal/slices"
	"github.com/notaryproject/notation/pkg/configutil"
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

type certRemoveOpts struct {
	names []string
}

// type certGenerateTestOpts struct {
// 	name      string
// 	bits      int
// 	trust     bool
// 	hosts     []string
// 	isDefault bool
// }

func certCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "certificate",
		Aliases: []string{"cert"},
		Short:   "Manage trust store and certificates used for verification",
	}

	// command.AddCommand(certAddCommand(nil), certListCommand(), certRemoveCommand(nil), certGenerateTestCommand(nil))
	command.AddCommand(certAddCommand(nil), certListCommand(nil), certShowCommand(nil), certRemoveCommand(nil))
	return command
}

func certAddCommand(opts *certAddOpts) *cobra.Command {
	if opts == nil {
		opts = &certAddOpts{}
	}
	command := &cobra.Command{
		Use:   "add -t type -s name path...",
		Short: "Add certificates to the trust store",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing certificate path")
			}
			opts.path = args
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return addCert(opts)
		},
	}
	command.Flags().StringVarP(&opts.storeType, "type", "t", "ca", "specify trust store type, options: ca, tsa")
	command.Flags().StringVarP(&opts.namedStore, "store", "s", "", "specify named store")
	return command
}

func certListCommand(opts *certListOpts) *cobra.Command {
	if opts == nil {
		opts = &certListOpts{}
	}
	command := &cobra.Command{
		Use:     "list [-t type] [-s name]",
		Aliases: []string{"ls"},
		Short:   "List certificates used for verification",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listCerts(opts)
		},
	}
	command.Flags().StringVarP(&opts.storeType, "type", "t", "", "specify trust store type, options: ca, tsa")
	command.Flags().StringVarP(&opts.namedStore, "store", "s", "", "specify named store")
	return command
}

func certShowCommand(opts *certShowOpts) *cobra.Command {
	if opts == nil {
		opts = &certShowOpts{}
	}
	command := &cobra.Command{
		Use:   "show -t type -s name -f fileName",
		Short: "Show certificate details given trust store type, named store, and cert file name",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCerts(opts)
		},
	}
	command.Flags().StringVarP(&opts.storeType, "type", "t", "", "specify trust store type, options: ca, tsa")
	command.Flags().StringVarP(&opts.namedStore, "store", "s", "", "specify named store")
	command.Flags().StringVarP(&opts.cert, "fileName", "f", "", "specify cert file name")
	return command
}

func certRemoveCommand(opts *certRemoveOpts) *cobra.Command {
	if opts == nil {
		opts = &certRemoveOpts{}
	}
	command := &cobra.Command{
		Use:     "remove [name]...",
		Aliases: []string{"rm"},
		Short:   "Remove certificate from the verification list",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing certificate names")
			}
			opts.names = args
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return removeCerts(opts)
		},
	}
	return command
}

// func certGenerateTestCommand(opts *certGenerateTestOpts) *cobra.Command {
// 	if opts == nil {
// 		opts = &certGenerateTestOpts{}
// 	}
// 	command := &cobra.Command{
// 		Use:   "generate-test [host]...",
// 		Short: "Generates a test RSA key and a corresponding self-generated certificate chain",
// 		Args: func(cmd *cobra.Command, args []string) error {
// 			if len(args) == 0 {
// 				return errors.New("missing certificate hosts")
// 			}
// 			opts.hosts = args
// 			return nil
// 		},
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			return generateTestCert(opts)
// 		},
// 	}

// 	command.Flags().StringVarP(&opts.name, "name", "n", "", "key and certificate name")
// 	command.Flags().IntVarP(&opts.bits, "bits", "b", 2048, "RSA key bits")
// 	command.Flags().BoolVar(&opts.trust, "trust", false, "add the generated certificate to the verification list")
// 	setKeyDefaultFlag(command.Flags(), &opts.isDefault)
// 	return command
// }

func addCert(opts *certAddOpts) error {
	storeType := opts.storeType
	if storeType == "" {
		return errors.New("missing trust store type")
	}
	namedStore := opts.namedStore
	if namedStore == "" {
		return errors.New("missing named store")
	}
	var success []string
	var failure []string
	var errorSlice []error
	for _, p := range opts.path {
		err := AddCertCore(p, storeType, namedStore, false)
		if err != nil {
			failure = append(failure, p)
			errorSlice = append(errorSlice, err)
		} else {
			success = append(success, p)
		}
	}
	if len(success) != 0 {
		fmt.Println("Successfully added following certificates into Trust Store:")
		for _, p := range success {
			fmt.Println(p)
		}
	}

	if len(failure) != 0 {
		fmt.Println("Failed to add following certificates into Trust Store:")

		for ind := range failure {
			fmt.Printf("%s, with error \"%s\"\n", failure[ind], errorSlice[ind])
		}
	}

	return nil
}

// AddCertCore adds a single cert file at path to the trust store with dir
// truststore/x509/storeType/namedStore
func AddCertCore(path, storeType, namedStore string, display bool) error {
	// initialize
	certPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if storeType == "" {
		return errors.New("missing trust store type")
	}
	if namedStore == "" {
		return errors.New("missing named store")
	}

	// check if the target path is a cert (support PEM and DER formats)
	if _, err := corex509.ReadCertificateFile(certPath); err != nil {
		return err
	}

	// core process
	trustStorePath := dir.Path.X509TrustStore(storeType, namedStore)
	// check if certificate already in the trust store
	if _, err := os.Stat(filepath.Join(trustStorePath, filepath.Base(certPath))); err == nil {
		return errors.New("certificate already exists in the Trust Store, try delete it and add again")
	}
	_, err = osutil.Copy(certPath, trustStorePath)
	if err != nil {
		return err
	}

	// write out
	if display {
		fmt.Println(filepath.Base(certPath))
	}
	return nil
}

func listCerts(opts *certListOpts) error {
	namedStore := opts.namedStore
	storeType := opts.storeType

	// trust store path, has to be exist to continue
	path, err := dir.Path.ConfigFS.GetPath(dir.TrustStoreDir, "x509")
	if err != nil {
		return err
	}

	if namedStore == "" {
		if storeType != "" {
			return errors.New("cannot only specify trust store type without named store")
		}
		// List all certificates in the trust store, display empty if there's
		// no certificate yet
		return checkError(certsPrinter(path))
	}

	if storeType == "" {
		// List all certificates under named store namedStore, display empty if
		// there's no such certificate
		path, err = dir.Path.ConfigFS.GetPath(dir.TrustStoreDir, "x509", "ca", namedStore)
		if err = checkError(err); err != nil {
			return err
		}
		if err = checkError(certsPrinter(path)); err != nil {
			return err
		}

		path, err = dir.Path.ConfigFS.GetPath(dir.TrustStoreDir, "x509", "tsa", namedStore)
		if err = checkError(err); err != nil {
			return err
		}
		if err = checkError(certsPrinter(path)); err != nil {
			return err
		}
	} else {
		// List all certificates under trust store type storeType and
		// named store namedStore, display empty if there's no such certificate
		path, err = dir.Path.ConfigFS.GetPath(dir.TrustStoreDir, "x509", storeType, namedStore)
		if err = checkError(err); err != nil {
			return err
		}
		if err = checkError(certsPrinter(path)); err != nil {
			return err
		}
	}
	return nil
}

func showCerts(opts *certShowOpts) error {
	storeType := opts.storeType
	if storeType == "" {
		return errors.New("missing trust store type")
	}
	namedStore := opts.namedStore
	if namedStore == "" {
		return errors.New("missing named store")
	}
	cert := opts.cert
	if cert == "" {
		return errors.New("missing cert fileName")
	}

	path, err := dir.Path.ConfigFS.GetPath(dir.TrustStoreDir, "x509", storeType, namedStore, cert)
	if err != nil {
		return err
	}
	certs, err := corex509.ReadCertificateFile(path)
	if err != nil || len(certs) == 0 {
		return err
	}
	showRootCA(certs[len(certs)-1])
	return nil
}

func removeCerts(opts *certRemoveOpts) error {
	// core process
	cfg, err := configutil.LoadConfigOnce()
	if err != nil {
		return err
	}

	var removedNames []string
	for _, name := range opts.names {
		idx := slices.Index(cfg.VerificationCertificates.Certificates, name)
		if idx < 0 {
			return errors.New(name + ": not found")
		}
		cfg.VerificationCertificates.Certificates = slices.Delete(cfg.VerificationCertificates.Certificates, idx)
		removedNames = append(removedNames, name)
	}
	if err := cfg.Save(); err != nil {
		return err
	}

	// write out
	for _, name := range removedNames {
		fmt.Println(name)
	}
	return nil
}

// certsPrinter walk through path as root and prints out all regular files in it
func certsPrinter(path string) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Mode().IsRegular() {
			fmt.Println(path)
		}
		return nil
	})
}

func showRootCA(cert *x509.Certificate) {
	fmt.Println("Issuer: ", cert.Issuer)
	fmt.Println("Subject: ", cert.Subject)
	fmt.Println("Valid from: ", cert.NotBefore)
	fmt.Println("Valid to: ", cert.NotAfter)
	fmt.Println("Version: ", cert.Version)
	fmt.Println("Serial number: ", cert.SerialNumber)
	fmt.Println("Signature Algorithm: ", cert.SignatureAlgorithm)
	fmt.Println("Public Key Algorithm: ", cert.PublicKeyAlgorithm)
	fmt.Println("Public Key: ", cert.PublicKey)
	keyUsage, ok := corex509.KeyUsageNameMap[cert.KeyUsage]
	if ok {
		fmt.Println("Key Usage: ", keyUsage)
	}
	var extKeyUsage []string
	for _, u := range cert.ExtKeyUsage {
		extKeyUsageString, ok := corex509.ExtKeyUsagesNameMap[u]
		if ok {
			extKeyUsage = append(extKeyUsage, extKeyUsageString)
		}
	}
	fmt.Println("Extended key usages: ", extKeyUsage)
	fmt.Println("Basic Constraints Valid: ", cert.BasicConstraintsValid)
	fmt.Println("IsCA: ", cert.IsCA)
}

func checkError(err error) error {
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}
