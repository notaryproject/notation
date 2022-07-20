package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/notaryproject/notation-go/plugin/manager"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/notaryproject/notation/internal/slices"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	keyDefaultFlag = &pflag.Flag{
		Name:      "default",
		Shorthand: "d",
		Usage:     "mark as default",
	}
	setKeyDefaultFlag = func(fs *pflag.FlagSet, p *bool) {
		fs.BoolVarP(p, keyDefaultFlag.Name, keyDefaultFlag.Shorthand, false, keyDefaultFlag.Usage)
	}
)

type keyAddOpts struct {
	name         string
	plugin       string
	id           string
	pluginConfig string
	isDefault    bool
	keyPath      string
	certPath     string
}

type keyUpdateOpts struct {
	name      string
	isDefault bool
}

type keyRemoveOpts struct {
	names []string
}

func keyCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "key",
		Short: "Manage keys used for signing",
	}
	command.AddCommand(keyAddCommand(nil), keyUpdateCommand(nil), keyListCommand(), keyRemoveCommand(nil))
	return command
}

func keyAddCommand(opts *keyAddOpts) *cobra.Command {
	if opts == nil {
		opts = &keyAddOpts{}
	}
	command := &cobra.Command{
		Use:   "add [key_path cert_path]",
		Short: "Add key to signing key list",
		Args:  cobra.MaximumNArgs(2),
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) >= 2 {
				opts.keyPath = args[0]
				opts.certPath = args[1]
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return addKey(cmd, opts)
		},
	}
	command.Flags().StringVarP(&opts.name, "name", "n", "", "key name")
	command.MarkFlagRequired("name")

	command.Flags().StringVarP(&opts.plugin, "plugin", "p", "", "signing plugin name")
	command.Flags().StringVar(&opts.id, "id", "", "key id (required if --plugin is set)")

	cmd.SetPflagPluginConfig(command.Flags(), &opts.pluginConfig)
	setKeyDefaultFlag(command.Flags(), &opts.isDefault)
	return command
}

func keyUpdateCommand(opts *keyUpdateOpts) *cobra.Command {
	if opts == nil {
		opts = &keyUpdateOpts{}
	}
	command := &cobra.Command{
		Use:     "update [name]",
		Aliases: []string{"set"},
		Short:   "Update key in signing key list",
		Args:    cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.name = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateKey(cmd, opts)
		},
	}

	setKeyDefaultFlag(command.Flags(), &opts.isDefault)
	return command
}

func keyListCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List keys used for signing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listKeys(cmd)
		},
	}
}

func keyRemoveCommand(opts *keyRemoveOpts) *cobra.Command {
	if opts == nil {
		opts = &keyRemoveOpts{}
	}
	return &cobra.Command{
		Use:     "remove [name]...",
		Aliases: []string{"rm"},
		Short:   "Remove key from signing key list",
		Args:    cobra.MinimumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.names = args
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return removeKeys(cmd, opts)
		},
	}
}

func addKey(command *cobra.Command, opts *keyAddOpts) error {
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}
	var key config.KeySuite
	pluginName := opts.plugin
	name := opts.name
	if pluginName != "" {
		key, err = addExternalKey(command.Context(), opts, pluginName, name)
	} else {
		key, err = newX509KeyPair(opts, name)
	}
	if err != nil {
		return err
	}

	isDefault := opts.isDefault
	err = addKeyCore(cfg, key, isDefault)
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

func addExternalKey(ctx context.Context, opts *keyAddOpts, pluginName, keyName string) (config.KeySuite, error) {
	id := opts.id
	if id == "" {
		return config.KeySuite{}, errors.New("missing key id")
	}
	mgr := manager.New(config.PluginDirPath)
	p, err := mgr.Get(ctx, pluginName)
	if err != nil {
		return config.KeySuite{}, err
	}
	if p.Err != nil {
		return config.KeySuite{}, fmt.Errorf("invalid plugin: %w", p.Err)
	}
	pluginConfig, err := cmd.ParseFlagPluginConfig(opts.pluginConfig)
	if err != nil {
		return config.KeySuite{}, err
	}
	return config.KeySuite{
		Name: keyName,
		ExternalKey: &config.ExternalKey{
			ID:           id,
			PluginName:   pluginName,
			PluginConfig: pluginConfig,
		},
	}, nil
}

func newX509KeyPair(opts *keyAddOpts, keyName string) (config.KeySuite, error) {
	if opts.keyPath == "" {
		return config.KeySuite{}, errors.New("missing key and certificate paths")
	}
	if opts.certPath == "" {
		return config.KeySuite{}, errors.New("missing certificate path for the corresponding key")
	}

	keyPath, err := filepath.Abs(opts.keyPath)
	if err != nil {
		return config.KeySuite{}, err
	}
	certPath, err := filepath.Abs(opts.certPath)
	if err != nil {
		return config.KeySuite{}, err
	}

	// check key / cert pair
	if _, err := tls.LoadX509KeyPair(certPath, keyPath); err != nil {
		return config.KeySuite{}, err
	}
	return config.KeySuite{
		Name:        keyName,
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

func updateKey(command *cobra.Command, opts *keyUpdateOpts) error {
	// initialize
	name := opts.name
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
	if !opts.isDefault {
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

func listKeys(command *cobra.Command) error {
	// core process
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}

	// write out
	return ioutil.PrintKeyMap(os.Stdout, cfg.SigningKeys.Default, cfg.SigningKeys.Keys)
}

func removeKeys(command *cobra.Command, opts *keyRemoveOpts) error {
	// initialize
	names := opts.names
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
