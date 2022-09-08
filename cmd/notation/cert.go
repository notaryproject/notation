package main

import (
	"crypto/x509"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	corex509 "github.com/notaryproject/notation-core-go/x509"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation/internal/osutil"
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
	storeType  string
	namedStore string
	cert       string
	all        bool
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
		Use:   "show -t type -s name fileName",
		Short: "Show certificate details given trust store type, named store, and cert file name. If input is a certificate chain, only details of the root certificate is displayed",
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
	command.Flags().StringVarP(&opts.storeType, "type", "t", "", "specify trust store type, options: ca, tsa")
	command.Flags().StringVarP(&opts.namedStore, "store", "s", "", "specify named store")
	return command
}

func certRemoveCommand(opts *certRemoveOpts) *cobra.Command {
	if opts == nil {
		opts = &certRemoveOpts{}
	}
	command := &cobra.Command{
		Use:     "delete [-t type] -s name {--all | fileName}",
		Aliases: []string{"rm"},
		Short:   "Delete certificate from the trust store",
		Args: func(cmd *cobra.Command, args []string) error {
			if !opts.all && len(args) == 0 {
				return errors.New("needs to specify certificate name or set --all flag")
			}
			if len(args) != 0 {
				opts.cert = args[0]
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return removeCerts(opts)
		},
	}
	command.Flags().StringVarP(&opts.storeType, "type", "t", "", "specify trust store type, options: ca, tsa")
	command.Flags().StringVarP(&opts.namedStore, "store", "s", "", "specify named store")
	command.Flags().BoolVarP(&opts.all, "all", "a", false, "if set to true, remove all certificates in the named store")
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
		return errors.New("certificate already exists in the Trust Store, delete it and try again")
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

	// List all certificates under truststore/x509, display empty if there's
	// no certificate yet
	if namedStore == "" && storeType == "" {
		return checkError(printCerts(path))
	}

	// List all certificates under truststore/x509/storeType/namedStore
	// display empty if there's no such certificate
	if namedStore != "" && storeType != "" {
		path, err = dir.Path.ConfigFS.GetPath(dir.TrustStoreDir, "x509", storeType, namedStore)
		if err = checkError(err); err != nil {
			return err
		}
		if err = checkError(printCerts(path)); err != nil {
			return err
		}

		return nil
	}

	// List all certificates under x509/storeType
	if storeType != "" {
		path, err = dir.Path.ConfigFS.GetPath(dir.TrustStoreDir, "x509", storeType)
		if err = checkError(err); err != nil {
			return err
		}
		if err = checkError(printCerts(path)); err != nil {
			return err
		}
	} else {
		// List all certificates under named store namedStore, display empty if
		// there's no such certificate
		path, err = dir.Path.ConfigFS.GetPath(dir.TrustStoreDir, "x509", "ca", namedStore)
		if err = checkError(err); err == nil {
			if err = checkError(printCerts(path)); err != nil {
				return err
			}
		}

		path, err = dir.Path.ConfigFS.GetPath(dir.TrustStoreDir, "x509", "tsa", namedStore)
		if err = checkError(err); err == nil {
			if err = checkError(printCerts(path)); err != nil {
				return err
			}
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
	if err != nil {
		return err
	}
	if len(certs) == 0 {
		return fmt.Errorf("missing certificate file under %s", path)
	}
	showRootCA(certs[len(certs)-1])

	return nil
}

func removeCerts(opts *certRemoveOpts) error {
	namedStore := opts.namedStore
	if namedStore == "" {
		return errors.New("missing named store")
	}
	storeType := opts.storeType
	var errorSlice []error

	if opts.all {
		if storeType == "" {
			// Delete all certificates under namedStore
			errorSlice = removeCertsCore("ca", namedStore, "", errorSlice)
			errorSlice = removeCertsCore("tsa", namedStore, "", errorSlice)
		} else {
			// Delete all certificates under storeType/namedStore
			errorSlice = removeCertsCore(storeType, namedStore, "", errorSlice)
		}
		if len(errorSlice) > 0 {
			fmt.Println("Failed to clear following named stores:")
			for _, err := range errorSlice {
				fmt.Println(err.Error())
			}
		}

		return nil
	}

	// Delete a certain certificate with path storeType/namedStore/cert
	if storeType == "" {
		return errors.New("missing trust store type")
	}
	if opts.cert == "" {
		return errors.New("missing certificate fileName")
	}
	errorSlice = removeCertsCore(storeType, namedStore, opts.cert, errorSlice)
	if len(errorSlice) > 0 {
		fmt.Println("Failed to delete following certificate:")
		for _, err := range errorSlice {
			fmt.Println(err.Error())
		}
	}

	return nil
}

func removeCertsCore(storeType, namedStore, cert string, errorSlice []error) []error {
	if cert == "" {
		path, err := dir.Path.ConfigFS.GetPath(dir.TrustStoreDir, "x509", storeType, namedStore)
		if err == nil {
			if err = osutil.CleanDir(path); err != nil {
				errorSlice = append(errorSlice, fmt.Errorf("%s with error \"%s\"", path, err.Error()))
			}
		} else {
			errorSlice = append(errorSlice, fmt.Errorf("%s with error \"%s\"", path, err.Error()))
		}
	} else {
		path, err := dir.Path.ConfigFS.GetPath(dir.TrustStoreDir, "x509", storeType, namedStore, cert)
		if err == nil {
			if err = os.RemoveAll(path); err != nil {
				errorSlice = append(errorSlice, fmt.Errorf("%s with error \"%s\"", path, err.Error()))
			} else {
				fmt.Printf("Successfully deleted %s\n", path)
				return []error{}
			}
		} else {
			errorSlice = append(errorSlice, fmt.Errorf("%s with error \"%s\"", path, err.Error()))
		}
	}

	return errorSlice
}

// printCerts walks through path and prints out all regular files in it
func printCerts(path string) error {
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

// showRootCA displays details of a root certificate
func showRootCA(cert *x509.Certificate) {
	fmt.Println("Issuer:", cert.Issuer)
	fmt.Println("Subject:", cert.Subject)
	fmt.Println("Valid from:", cert.NotBefore)
	fmt.Println("Valid to:", cert.NotAfter)
	fmt.Println("Version:", cert.Version)
	fmt.Println("Serial number:", cert.SerialNumber)
	fmt.Println("Signature Algorithm:", cert.SignatureAlgorithm)
	fmt.Println("Public Key Algorithm:", cert.PublicKeyAlgorithm)
	fmt.Println("Public Key:", cert.PublicKey)

	// KeyUsage
	var keyUsage []string
	for k, v := range corex509.KeyUsageNameMap {
		if cert.KeyUsage&k != 0 {
			keyUsage = append(keyUsage, v)
		}
	}
	keyUsagePrint := strings.Join(keyUsage, ", ")
	fmt.Println("Key Usage:", keyUsagePrint)

	// ExtKeyUsage
	var extKeyUsage []string
	for _, u := range cert.ExtKeyUsage {
		extKeyUsageString, ok := corex509.ExtKeyUsagesNameMap[u]
		if ok {
			extKeyUsage = append(extKeyUsage, extKeyUsageString)
		}
	}
	extKeyUsagePrint := strings.Join(extKeyUsage, ", ")
	fmt.Println("Extended key usages:", extKeyUsagePrint)

	fmt.Println("Basic Constraints Valid:", cert.BasicConstraintsValid)
	fmt.Println("IsCA:", cert.IsCA)
}

func checkError(err error) error {
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}
