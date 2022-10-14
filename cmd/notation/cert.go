package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/notaryproject/notation-core-go/x509"
	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/notaryproject/notation/internal/slices"
	"github.com/notaryproject/notation/pkg/configutil"
	"github.com/spf13/cobra"
)

type certAddOpts struct {
	path string
	name string
}

type certRemoveOpts struct {
	names []string
}

type certGenerateTestOpts struct {
	name      string
	bits      int
	trust     bool
	hosts     []string
	isDefault bool
}

func certCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "certificate",
		Aliases: []string{"cert"},
		Short:   "Manage certificates used for verification",
		Long: `Manage certificates used for verification

Example - Generate a self-signed test certificate and add it to the verification list:
  notation cert generate-test -n localcert --trust

Example - Add the certificate to the trusted certificate list:
  notation cert add --name <key_name> <cert_path>

Example - Generate a self-signed test certificate and set it to default:
  notation cert generate-test -d <key_name>

Example - List trusted certificates:
  notation cert ls

Example - Remove the certificate from the trusted certificate list:
  notation cert rm <key_name>
`,
	}

	command.AddCommand(certAddCommand(nil), certListCommand(), certRemoveCommand(nil), certGenerateTestCommand(nil))
	return command
}

func certAddCommand(opts *certAddOpts) *cobra.Command {
	if opts == nil {
		opts = &certAddOpts{}
	}
	command := &cobra.Command{
		Use:   "add [path]",
		Short: "Add certificate to verification list",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing certificate path")
			}
			opts.path = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return addCert(opts)
		},
	}
	command.Flags().StringVarP(&opts.name, "name", "n", "", "certificate name")
	return command
}

func certListCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List certificates used for verification",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listCerts()
		},
	}
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
func certGenerateTestCommand(opts *certGenerateTestOpts) *cobra.Command {
	if opts == nil {
		opts = &certGenerateTestOpts{}
	}
	command := &cobra.Command{
		Use:   "generate-test [host]...",
		Short: "Generates a test RSA key and a corresponding self-generated certificate chain",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing certificate hosts")
			}
			opts.hosts = args
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateTestCert(opts)
		},
	}

	command.Flags().StringVarP(&opts.name, "name", "n", "", "key and certificate name")
	command.Flags().IntVarP(&opts.bits, "bits", "b", 2048, "RSA key bits")
	command.Flags().BoolVar(&opts.trust, "trust", false, "add the generated certificate to the verification list")
	setKeyDefaultFlag(command.Flags(), &opts.isDefault)
	return command
}

func addCert(opts *certAddOpts) error {
	// initialize
	path, err := filepath.Abs(opts.path)
	if err != nil {
		return err
	}
	name := opts.name

	// check if the target path is a cert
	if _, err := x509.ReadCertificateFile(path); err != nil {
		return err
	}

	// core process
	cfg, err := configutil.LoadConfigOnce()
	if err != nil {
		return err
	}
	if err := addCertCore(cfg, name, path); err != nil {
		return err
	}
	if err := cfg.Save(); err != nil {
		return err
	}

	// write out
	fmt.Println(name)
	return nil
}

func addCertCore(cfg *config.Config, name, path string) error {
	if slices.Contains(cfg.VerificationCertificates.Certificates, name) {
		return errors.New(name + ": already exists")
	}
	cfg.VerificationCertificates.Certificates = append(cfg.VerificationCertificates.Certificates, config.CertificateReference{
		Name: name,
		Path: path,
	})
	return nil
}

func listCerts() error {
	// core process
	cfg, err := configutil.LoadConfigOnce()
	if err != nil {
		return err
	}

	// write out
	return ioutil.PrintCertificateMap(os.Stdout, cfg.VerificationCertificates.Certificates)
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
