package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/notaryproject/notation-go-lib/crypto/cryptoutil"
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
			&cli.StringFlag{
				Name:    "id",
				Aliases: []string{"i"},
				Usage:   "key id",
			},
			&cli.StringFlag{
				Name:    "plugin",
				Aliases: []string{"p"},
				Usage:   "key plugin",
			},
			&cli.BoolFlag{
				Name:  "kms",
				Usage: "add key to KMS cert list",
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
	if ctx.Bool("kms") {
		return addKMSCertToList(ctx)
	}
	return addCertToList(ctx)
}

func addCertToList(ctx *cli.Context) error {
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
		name = nameFromPath(path)
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

func addKMSCertToList(ctx *cli.Context) error {
	keyID := ctx.String("id")
	plugin := ctx.String("plugin")
	name := ctx.String("name")

	// initialize
	if keyID == "" {
		return errors.New("missing key id")
	}
	if plugin == "" {
		return errors.New("missing key plugin")
	}
	if name == "" {
		name = keyID
	}

	// core process
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}
	if err = addKMSCertCore(cfg, name, keyID, plugin); err != nil {
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
	if ok := cfg.VerificationCertificates.Certificates.Append(name, path); !ok {
		return errors.New(name + ": already exists")
	}
	return nil
}

func addKMSCertCore(cfg *config.File, name, keyID, plugin string) error {
	if ok := cfg.VerificationCertificates.KMSCerts.Append(name, keyID, plugin); !ok {
		return errors.New(name + ": already exists")
	}
	return nil
}

func listCerts(ctx *cli.Context) error {
	// core process
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}

	// write out
	printCertificateSet(cfg.VerificationCertificates.Certificates, cfg.VerificationCertificates.KMSCerts)
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
			if ok := cfg.VerificationCertificates.KMSCerts.Remove(name); !ok {
				return errors.New(name + ": not found")
			}
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

func printCertificateSet(s config.CertificateMap, k config.KMSProfileMap) {
	maxNameSize := 0
	for _, ref := range s {
		if len(ref.Name) > maxNameSize {
			maxNameSize = len(ref.Name)
		}
	}
	format := fmt.Sprintf("%%-%ds\t%%s\n", maxNameSize)
	fmt.Printf(format, "NAME", "PATH")
	for _, ref := range s {
		fmt.Printf(format, ref.Name, ref.Path)
	}

	var maxKeyIDSize int
	for _, ref := range k {
		if len(ref.Name) > maxNameSize {
			maxNameSize = len(ref.Name)
		}
		if len(ref.ID) > maxKeyIDSize {
			maxKeyIDSize = len(ref.ID)
		}
	}
	fmt.Println()
	// iterate over KMS profiles
	format = fmt.Sprintf("%%-%ds\t%%-%ds\t%%s\n", maxNameSize, maxKeyIDSize)
	fmt.Printf(format, "NAME", "ID", "PLUGIN NAME")
	for _, ref := range k {
		fmt.Printf(format, ref.Name, ref.ID, ref.PluginName)
	}
}

func nameFromPath(path string) string {
	base := filepath.Base(path)
	name := base[:len(base)-len(filepath.Ext(base))]
	if name == "" {
		return base
	}
	return name
}
