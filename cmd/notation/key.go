package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"path/filepath"

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
		ArgsUsage: "<key_path> <cert_path>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "key name",
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
				Usage: "add key to KMS key list",
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
	if ctx.Bool("kms") {
		return addKMSKeyToList(ctx)
	}

	return addKeyToList(ctx)
}

func addKeyToList(ctx *cli.Context) error {
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
		name = nameFromPath(keyPath)
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
	isDefault, err := addKeyCore(cfg, name, keyPath, certPath, ctx.Bool(keyDefaultFlag.Name))
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

func addKMSKeyToList(ctx *cli.Context) error {
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
	isDefault, err := addKMSKeyCore(cfg, name, keyID, plugin, ctx.Bool(keyDefaultFlag.Name))
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

func addKeyCore(cfg *config.File, name, keyPath, certPath string, markDefault bool) (bool, error) {
	if ok := cfg.SigningKeys.Keys.Append(name, keyPath, certPath); !ok {
		return false, errors.New(name + ": already exists")
	}
	if markDefault {
		cfg.SigningKeys.Default = name
	}
	return cfg.SigningKeys.Default == name, nil
}

func addKMSKeyCore(cfg *config.File, name, keyID, plugin string, markDefault bool) (bool, error) {
	if ok := cfg.SigningKeys.KMSKeys.Append(name, keyID, plugin); !ok {
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
	printKeySet(cfg.SigningKeys.Default, cfg.SigningKeys.Keys, cfg.SigningKeys.KMSKeys)
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
			if ok := cfg.SigningKeys.KMSKeys.Remove(name); !ok {
				return errors.New(name + ": not found")
			}
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

func printKeySet(target string, s config.KeyMap, k config.KMSProfileMap) {
	if len(s) == 0 && len(k) == 0 {
		fmt.Println("NAME\tPATH")
		return
	}

	var maxNameSize, maxKeyPathSize int
	for _, ref := range s {
		if len(ref.Name) > maxNameSize {
			maxNameSize = len(ref.Name)
		}
		if len(ref.KeyPath) > maxKeyPathSize {
			maxKeyPathSize = len(ref.KeyPath)
		}
	}
	format := fmt.Sprintf("%%c %%-%ds\t%%-%ds\t%%s\n", maxNameSize, maxKeyPathSize)
	fmt.Printf(format, ' ', "NAME", "KEY PATH", "CERTIFICATE PATH")
	for _, ref := range s {
		mark := ' '
		if ref.Name == target {
			mark = '*'
		}
		fmt.Printf(format, mark, ref.Name, ref.KeyPath, ref.CertificatePath)
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
	format = fmt.Sprintf("%%c %%-%ds\t%%-%ds\t%%s\n", maxNameSize, maxKeyIDSize)
	fmt.Printf(format, ' ', "NAME", "ID", "PLUGIN NAME")
	for _, ref := range k {
		mark := ' '
		if ref.Name == target {
			mark = '*'
		}
		fmt.Printf(format, mark, ref.Name, ref.ID, ref.PluginName)
	}
}
