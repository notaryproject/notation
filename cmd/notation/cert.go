package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/notaryproject/notation-go-lib/crypto/cryptoutil"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/notaryproject/notation/pkg/signature"
	"github.com/notaryproject/notation/pkg/test"
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

func generateTestCert(ctx *cli.Context) error {
	// initialize
	hosts := ctx.Args().Slice()
	if len(hosts) == 0 {
		return errors.New("missing certificate hosts")
	}
	name := ctx.String("name")
	if name == "" {
		name = hosts[0]
	}

	// generate RSA private key
	bits := ctx.Int("bits")
	fmt.Println("generating RSA Key with", bits, "bits")
	key, keyBytes, err := test.GenerateTestKey(bits)
	if err != nil {
		return err
	}

	// generate self-signed certificate
	cert, certBytes, err := test.GenerateTestSelfSignedCert(key, hosts, ctx.Duration("expiry"))
	if err != nil {
		return err
	}
	fmt.Println("generated certificates expiring on", cert.NotAfter.Format(time.RFC3339))

	// write private key
	keyPath := config.KeyPath(name)
	if err := osutil.WriteFileWithPermission(keyPath, keyBytes, 0600, false); err != nil {
		return fmt.Errorf("failed to write key file: %v", err)
	}
	fmt.Println("wrote key:", keyPath)

	// write self-signed certificate
	certPath := config.CertificatePath(name)
	if err := osutil.WriteFileWithPermission(certPath, certBytes, 0644, false); err != nil {
		return fmt.Errorf("failed to write certificate file: %v", err)
	}
	fmt.Println("wrote certificate:", certPath)

	// update config
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}
	isDefaultKey, err := signature.AddKeyCore(cfg, name, keyPath, certPath, true)
	if err != nil {
		return err
	}
	trust := ctx.Bool("trust")
	if trust {
		if err := signature.AddCertCore(cfg, name, certPath); err != nil {
			return err
		}
	}
	if err := cfg.Save(); err != nil {
		return err
	}

	// write out
	fmt.Printf("%s: added to the key list\n", name)
	if isDefaultKey {
		fmt.Printf("%s: marked as default\n", name)
	}
	if trust {
		fmt.Printf("%s: added to the certificate list\n", name)
	}
	return nil
}

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
	if name == "" {
		name = signature.NameFromPath(path)
	}

	// check if the target path is a cert
	if _, err := cryptoutil.ReadCertificateFile(path); err != nil {
		return err
	}

	// core process
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}
	if err := signature.AddCertCore(cfg, name, path); err != nil {
		return err
	}
	if err := cfg.Save(); err != nil {
		return err
	}

	// write out
	fmt.Println(name)
	return nil
}

func listCerts(ctx *cli.Context) error {
	// core process
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}

	// write out
	signature.PrintCertificateSet(cfg.VerificationCertificates.Certificates)
	return nil
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
		if ok := cfg.VerificationCertificates.Certificates.Remove(name); !ok {
			return errors.New(name + ": not found")
		}
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
