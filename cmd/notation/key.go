package main

import (
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
	setKeyDefaultFlag = func(command *cobra.Command) {
		command.Flags().BoolP(keyDefaultFlag.Name, keyDefaultFlag.Shorthand, false, keyDefaultFlag.Usage)
	}
)

func keyCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "key",
		Short: "Manage keys used for signing",
	}
	command.AddCommand(keyAddCommand(), keyUpdateCommand(), keyListCommand(), keyRemoveCommand())
	return command
}

func keyAddCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "add [key_path cert_path]",
		Short: "Add key to signing key list",
		RunE: func(cmd *cobra.Command, args []string) error {
			return addKey(cmd)
		},
	}
	// TODO: test required
	command.Flags().StringP("name", "n", "", "key name")
	command.MarkFlagRequired("name")

	command.Flags().StringP("plugin", "p", "", "signing plugin name")
	command.Flags().String("id", "", "key id (required if --plugin is set)")

	cmd.SetFlagPluginConfig(command)
	setKeyDefaultFlag(command)

	return command
}

func keyUpdateCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "update [name]",
		Aliases: []string{"set"},
		Short:   "Update key in signing key list",
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateKey(cmd)
		},
	}

	setKeyDefaultFlag(command)
	return command
}

func keyListCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List keys used for signing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listKeys(cmd)
		},
	}
	return command
}

func keyRemoveCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "remove [name]...",
		Aliases: []string{"rm"},
		Short:   "Remove key from signing key list",
		RunE: func(cmd *cobra.Command, args []string) error {
			return removeKeys(cmd)
		},
	}
	return command
}

func addKey(command *cobra.Command) error {
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}
	var key config.KeySuite
	pluginName, _ := command.Flags().GetString("plugin")
	name, _ := command.Flags().GetString("name")
	if pluginName != "" {
		key, err = addExternalKey(command, pluginName, name)
	} else {
		key, err = newX509KeyPair(command, name)
	}
	if err != nil {
		return err
	}

	isDefault, _ := command.Flags().GetBool(keyDefaultFlag.Name)
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

func addExternalKey(command *cobra.Command, pluginName, keyName string) (config.KeySuite, error) {
	id, _ := command.Flags().GetString("id")
	if id == "" {
		return config.KeySuite{}, errors.New("missing key id")
	}
	mgr := manager.New(config.PluginDirPath)
	p, err := mgr.Get(command.Context(), pluginName)
	if err != nil {
		return config.KeySuite{}, err
	}
	if p.Err != nil {
		return config.KeySuite{}, fmt.Errorf("invalid plugin: %w", p.Err)
	}
	pluginConfig, err := cmd.ParseFlagPluginConfig(command)
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

func newX509KeyPair(command *cobra.Command, keyName string) (config.KeySuite, error) {
	switch command.Flags().NArg() {
	case 0:
		return config.KeySuite{}, errors.New("missing key and certificate paths")
	case 1:
		return config.KeySuite{}, errors.New("missing certificate path for the corresponding key")
	}

	keyPath, err := filepath.Abs(command.Flags().Arg(0))
	if err != nil {
		return config.KeySuite{}, err
	}
	certPath, err := filepath.Abs(command.Flags().Arg(1))
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

func updateKey(command *cobra.Command) error {
	// initialize
	name := command.Flags().Arg(0)
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
	if isDefault, _ := command.Flags().GetBool(keyDefaultFlag.Name); !isDefault {
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

func removeKeys(command *cobra.Command) error {
	// initialize
	names := command.Flags().Args()
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
