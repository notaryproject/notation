package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/notaryproject/notation/internal/slices"
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
		ArgsUsage: "[<key_path> <cert_path>]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "key name (required if --plugin is set)",
			},
			&cli.StringFlag{
				Name:  "id",
				Usage: "key id (required if --plugin is set)",
			},
			&cli.StringFlag{
				Name:    "plugin",
				Aliases: []string{"p"},
				Usage:   "signing plugin name",
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
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}
	var key *config.KeySuite
	pluginName := ctx.String("plugin")
	if pluginName != "" {
		key, err = addExternalKey(ctx, pluginName)
	} else {
		key, err = newX509KeyPair(ctx)
	}
	if err != nil {
		return err
	}

	isDefault := ctx.Bool(keyDefaultFlag.Name)
	err = addKeyCore(cfg, *key, isDefault)
	if err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	// write out
	if isDefault {
		fmt.Printf("%s: marked as default\n", key.Name)
	} else {
		fmt.Println(key.Name)
	}
	return nil
}

func addExternalKey(ctx *cli.Context, pluginName string) (*config.KeySuite, error) {
	name := ctx.String("name")
	if name == "" {
		return nil, errors.New("missing key name")
	}
	id := ctx.String("id")
	if id == "" {
		return nil, errors.New("missing key id")
	}
	mgr, err := plugin.NewManager()
	if err != nil {
		return nil, err
	}
	p, err := mgr.Get(pluginName)
	if err != nil {
		return nil, err
	}
	if p.Err != nil {
		return nil, fmt.Errorf("invalid plugin: %w", p.Err)
	}
	return &config.KeySuite{
		Name:        name,
		ExternalKey: &config.ExternalKey{ID: id, PluginName: pluginName},
	}, nil
}

func newX509KeyPair(ctx *cli.Context) (*config.KeySuite, error) {
	args := ctx.Args()
	switch args.Len() {
	case 0:
		return nil, errors.New("missing key and certificate paths")
	case 1:
		return nil, errors.New("missing certificate path for the corresponding key")
	}

	keyPath, err := filepath.Abs(args.Get(0))
	if err != nil {
		return nil, err
	}
	certPath, err := filepath.Abs(args.Get(1))
	if err != nil {
		return nil, err
	}
	name := ctx.String("name")
	if name == "" {
		name = nameFromPath(keyPath)
	}

	// check key / cert pair
	if _, err := tls.LoadX509KeyPair(certPath, keyPath); err != nil {
		return nil, err
	}
	return &config.KeySuite{
		Name:        name,
		X509KeyPair: &config.X509KeyPair{KeyPath: keyPath, CertificatePath: certPath},
	}, nil
}

func addKeyCore(cfg *config.File, key config.KeySuite, markDefault bool) error {
	if slices.Contains(cfg.SigningKeys.Keys, key.Name) {
		return errors.New(key.Name + ": already exists")
	}
	cfg.SigningKeys.Keys = append(cfg.SigningKeys.Keys, key)
	if markDefault {
		cfg.SigningKeys.Default = key.Name
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
	if !slices.Contains(cfg.SigningKeys.Keys, name) {
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
	ioutil.PrintKeyMap(os.Stdout, cfg.SigningKeys.Default, cfg.SigningKeys.Keys)
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
		idx := slices.Index(cfg.SigningKeys.Keys, name)
		if idx < 0 {
			return errors.New(name + ": not found")
		}
		cfg.SigningKeys.Keys = slices.Delete(cfg.SigningKeys.Keys, idx)
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
