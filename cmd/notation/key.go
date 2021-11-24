package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/notaryproject/notation/pkg/config"
	"github.com/notaryproject/notation/pkg/signature"
	"github.com/urfave/cli/v2"
)

var (
	keyCommand = &cli.Command{
		Name:  "key",
		Usage: "Manage keys used for signing",
		Subcommands: []*cli.Command{
			keyAddCommand,
			keyUpdateCommand,
			keyListCommand,
			keyRemoveCommand,
		},
	}

	keyDefaultFlag = &cli.BoolFlag{
		Name:    "default",
		Aliases: []string{"d"},
		Usage:   "mark as default",
	}

	keyAddCommand = &cli.Command{
		Name:      "add",
		Usage:     "Add key to signing key list",
		ArgsUsage: "<key_path> <cert_path>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "key name",
			},
			keyDefaultFlag,
		},
		Action: addKey,
	}

	keyUpdateCommand = &cli.Command{
		Name:      "update",
		Usage:     "Update key in signing key list",
		Aliases:   []string{"set"},
		ArgsUsage: "<name>",
		Flags: []cli.Flag{
			keyDefaultFlag,
		},
		Action: updateKey,
	}

	keyListCommand = &cli.Command{
		Name:    "list",
		Usage:   "List keys used for signing",
		Aliases: []string{"ls"},
		Action:  listKeys,
	}

	keyRemoveCommand = &cli.Command{
		Name:      "remove",
		Usage:     "Remove key from signing key list",
		Aliases:   []string{"rm"},
		ArgsUsage: "[name] ...",
		Action:    removeKeys,
	}
)

func addKey(ctx *cli.Context) error {
	// initialize
	args := ctx.Args()
	switch args.Len() {
	case 0:
		return errors.New("missing key and certificate paths")
	case 1:
		return errors.New("missing certificate path for the correspoding key")
	}

	keyPath, err := filepath.Abs(args.Get(0))
	if err != nil {
		return err
	}
	certPath, err := filepath.Abs(args.Get(1))
	if err != nil {
		return err
	}
	name := ctx.String("name")
	if name == "" {
		name = signature.NameFromPath(keyPath)
	}

	// check key / cert pair
	if _, err := tls.LoadX509KeyPair(certPath, keyPath); err != nil {
		return err
	}

	// core process
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}
	isDefault, err := signature.AddKeyCore(cfg, name, keyPath, certPath, ctx.Bool(keyDefaultFlag.Name))
	if err != nil {
		return err
	}
	if err := cfg.Save(); err != nil {
		return err
	}

	// write out
	if isDefault {
		fmt.Printf("%s: marked as default\n", name)
	} else {
		fmt.Println(name)
	}
	return nil
}

func updateKey(ctx *cli.Context) error {
	// initialize
	name := ctx.Args().First()
	if name == "" {
		return errors.New("missing key name")
	}

	// core process
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}
	if _, _, ok := cfg.SigningKeys.Keys.Get(name); !ok {
		return errors.New(name + ": not found")
	}
	if !ctx.Bool(keyDefaultFlag.Name) {
		return nil
	}
	if cfg.SigningKeys.Default != name {
		cfg.SigningKeys.Default = name
		if err := cfg.Save(); err != nil {
			return err
		}
	}

	// write out
	fmt.Printf("%s: marked as default\n", name)
	return nil
}

func listKeys(ctx *cli.Context) error {
	// core process
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}

	// write out
	signature.PrintKeySet(cfg.SigningKeys.Default, cfg.SigningKeys.Keys)
	return nil
}

func removeKeys(ctx *cli.Context) error {
	// initialize
	names := ctx.Args().Slice()
	if len(names) == 0 {
		return errors.New("missing key names")
	}

	// core process
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}

	prevDefault := cfg.SigningKeys.Default
	var removedNames []string
	for _, name := range names {
		if ok := cfg.SigningKeys.Keys.Remove(name); !ok {
			return errors.New(name + ": not found")
		}
		removedNames = append(removedNames, name)
		if prevDefault == name {
			cfg.SigningKeys.Default = ""
		}
	}
	if err := cfg.Save(); err != nil {
		return err
	}

	// write out
	for _, name := range removedNames {
		if prevDefault == name {
			fmt.Printf("%s: unmarked as default\n", name)
		} else {
			fmt.Println(name)
		}
	}
	return nil
}
