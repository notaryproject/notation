package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/notaryproject/notation-go-lib/signature/x509"
	"github.com/notaryproject/notation/pkg/config"
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
		ArgsUsage: "<path>",
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
	path := ctx.Args().First()
	if path == "" {
		return errors.New("missing key path")
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	name := ctx.String("name")
	if name == "" {
		name = nameFromPath(path)
	}

	// check if the target path is a key
	if _, err := x509.ReadPrivateKeyFile(path); err != nil {
		return err
	}

	// core process
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}
	isDefault, err := addKeyCore(cfg, name, path, ctx.Bool(keyDefaultFlag.Name))
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

func addKeyCore(cfg *config.File, name, path string, markDefault bool) (bool, error) {
	if ok := cfg.SigningKeys.Keys.Append(name, path); !ok {
		return false, errors.New(name + ": already exists")
	}
	if markDefault {
		cfg.SigningKeys.Default = name
	}
	return cfg.SigningKeys.Default == name, nil
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
	if _, ok := cfg.SigningKeys.Keys.Get(name); !ok {
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
	printKeySet(cfg.SigningKeys.Default, cfg.SigningKeys.Keys)
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

func printKeySet(target string, s config.FileSet) {
	if len(s) == 0 {
		fmt.Println("NAME\tPATH")
		return
	}

	maxNameSize := 0
	for _, ref := range s {
		if len(ref.Name) > maxNameSize {
			maxNameSize = len(ref.Name)
		}
	}
	format := fmt.Sprintf("%%c %%-%ds\t%%s\n", maxNameSize)
	fmt.Printf(format, ' ', "NAME", "PATH")
	for _, ref := range s {
		mark := ' '
		if ref.Name == target {
			mark = '*'
		}
		fmt.Printf(format, mark, ref.Name, ref.Path)
	}
}
