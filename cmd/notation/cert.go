package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/notaryproject/notation-core-go/x509"
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
	command.AddCommand(certAddCommand(nil), certListCommand(nil), certRemoveCommand(nil))
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
	for _, p := range opts.path {
		// initialize
		certPath, err := filepath.Abs(p)
		if err != nil {
			return err
		}

		// check if the target path is a cert (support PEM and DER formats)
		if _, err := x509.ReadCertificateFile(certPath); err != nil {
			continue
		}

		// core process
		path := dir.Path.X509TrustStore(storeType, namedStore)
		_, err = osutil.Copy(certPath, path)
		if err != nil {
			return err
		}

		// write out
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

func checkError(err error) error {
	// if path does not exist, the path can be used to create file
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}
