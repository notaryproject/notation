package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/notaryproject/notation-core-go/x509"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/notaryproject/notation/internal/slices"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/spf13/cobra"
)

func certCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "certificate",
		Aliases: []string{"cert"},
		Short:   "Manage certificates used for verification",
	}

	command.AddCommand(certAddCommand(), certListCommand(), certRemoveCommand(), certGenerateTestCommand())
	return command
}

func certAddCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "add [path]",
		Short: "Add certificate to verification list",
		RunE: func(cmd *cobra.Command, args []string) error {
			return addCert(cmd)
		},
	}
	command.Flags().StringP("name", "n", "", "certificate name")
	return command
}

func certListCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List certificates used for verification",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listCerts(cmd)
		},
	}
	return command
}
func certRemoveCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "remove [name]...",
		Aliases: []string{"rm"},
		Short:   "Remove certificate from the verification list",
		RunE: func(cmd *cobra.Command, args []string) error {
			return removeCerts(cmd)
		},
	}
	return command
}
func certGenerateTestCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "generate-test [host]...",
		Short: "Generates a test RSA key and a corresponding self-signed certificate",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateTestCert(cmd)
		},
	}
	command.Flags().StringP("name", "n", "", "key and certificate name")
	command.Flags().IntP("bits", "b", 2048, "RSA key bits")
	command.Flags().DurationP("expiry", "e", 365*24*time.Hour, "certificate expiry")
	command.Flags().Bool("trust", false, "add the generated certificate to the verification list")
	setKeyDefaultFlag(command)
	return command
}

func addCert(command *cobra.Command) error {
	// initialize

	path := command.Flags().Arg(0)
	if path == "" {
		return errors.New("missing certificate path")
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	name, _ := command.Flags().GetString("name")

	// check if the target path is a cert
	if _, err := x509.ReadCertificateFile(path); err != nil {
		return err
	}

	// core process
	cfg, err := config.LoadOrDefault()
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

func addCertCore(cfg *config.File, name, path string) error {
	if slices.Contains(cfg.VerificationCertificates.Certificates, name) {
		return errors.New(name + ": already exists")
	}
	cfg.VerificationCertificates.Certificates = append(cfg.VerificationCertificates.Certificates, config.CertificateReference{
		Name: name,
		Path: path,
	})
	return nil
}

func listCerts(command *cobra.Command) error {
	// core process
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}

	// write out
	return ioutil.PrintCertificateMap(os.Stdout, cfg.VerificationCertificates.Certificates)
}

func removeCerts(command *cobra.Command) error {
	// initialize
	names := command.Flags().Args()
	if len(names) == 0 {
		return errors.New("missing certificate names")
	}

	// core process
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}

	var removedNames []string
	for _, name := range names {
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
