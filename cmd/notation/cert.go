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
	"github.com/urfave/cli/v2"
)

var (
	certCommand = &cli.Command{
		Name:    "certificate",
		Aliases: []string{"cert"},
		Usage:   "Manage certificates used for verification",
		Subcommands: []*cli.Command{
			certAddCommand,
			certListCommand,
			certRemoveCommand,
			certGenerateTestCommand,
		},
	}

	certAddCommand = &cli.Command{
		Name:      "add",
		Usage:     "Add certificate to verification list",
		ArgsUsage: "<path>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "certificate name",
			},
		},
		Action: addCert,
	}

	certListCommand = &cli.Command{
		Name:    "list",
		Usage:   "List certificates used for verification",
		Aliases: []string{"ls"},
		Action:  listCerts,
	}

	certRemoveCommand = &cli.Command{
		Name:      "remove",
		Usage:     "Remove certificate from the verification list",
		Aliases:   []string{"rm"},
		ArgsUsage: "<name> ...",
		Action:    removeCerts,
	}

	certGenerateTestCommand = &cli.Command{
		Name:      "generate-test",
		Usage:     "Generates a test RSA key and a corresponding self-signed certificate",
		ArgsUsage: "<host> ...",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "key and certificate name",
			},
			&cli.IntFlag{
				Name:    "bits",
				Usage:   "RSA key bits",
				Aliases: []string{"b"},
				Value:   2048,
			},
			&cli.DurationFlag{
				Name:    "expiry",
				Aliases: []string{"e"},
				Usage:   "certificate expiry",
				Value:   365 * 24 * time.Hour,
			},
			&cli.BoolFlag{
				Name:  "trust",
				Usage: "add the generated certificate to the verification list",
			},
			keyDefaultFlag,
		},
		Action: generateTestCert,
	}
)

func addCert(ctx *cli.Context) error {
	// initialize
	path := ctx.Args().First()
	if path == "" {
		return errors.New("missing certificate path")
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	name := ctx.String("name")

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

func listCerts(ctx *cli.Context) error {
	// core process
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}

	// write out
	return ioutil.PrintCertificateMap(os.Stdout, cfg.VerificationCertificates.Certificates)
}

func removeCerts(ctx *cli.Context) error {
	// initialize
	names := ctx.Args().Slice()
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
