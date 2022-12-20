package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/log"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/notaryproject/notation/internal/slices"
	"github.com/notaryproject/notation/pkg/configutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	keyDefaultFlag = &pflag.Flag{
		Name:  "default",
		Usage: "mark as default",
	}
	setKeyDefaultFlag = func(fs *pflag.FlagSet, p *bool) {
		fs.BoolVarP(p, keyDefaultFlag.Name, keyDefaultFlag.Shorthand, false, keyDefaultFlag.Usage)
	}
)

type keyAddOpts struct {
	cmd.LoggingFlagOpts
	name         string
	plugin       string
	id           string
	pluginConfig []string
	isDefault    bool
}

type keyUpdateOpts struct {
	cmd.LoggingFlagOpts
	name      string
	isDefault bool
}

type keyDeleteOpts struct {
	cmd.LoggingFlagOpts
	names []string
}

func keyCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "key",
		Short: "Manage keys used for signing",
		Long: `Manage keys used for signing

Example - Add a key to signing key list:
  notation key add --plugin <plugin_name> --id <key_id> <key_name>

Example - List keys used for signing:
  notation key ls

Example - Update the default signing key:
  notation key set --default <key_name>

Example - Delete the key from signing key list:
  notation key delete <key_name>...
`,
	}
	command.AddCommand(keyAddCommand(nil), keyUpdateCommand(nil), keyListCommand(), keyDeleteCommand(nil))

	return command
}

func keyAddCommand(opts *keyAddOpts) *cobra.Command {
	if opts == nil {
		opts = &keyAddOpts{}
	}
	command := &cobra.Command{
		Use:   "add --plugin <plugin_name> [flags] <key_name>",
		Short: "Add key to signing key list",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("either missing key name or unnecessary parameters passed")
			}
			opts.name = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return addKey(cmd.Context(), opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	command.Flags().StringVarP(&opts.plugin, "plugin", "p", "", "signing plugin name")
	command.MarkFlagRequired("plugin")

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
		Use:     "update [flags] <key_name>",
		Aliases: []string{"set"},
		Short:   "Update key in signing key list",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing key name")
			}
			opts.name = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateKey(cmd.Context(), opts)
		},
	}

	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	setKeyDefaultFlag(command.Flags(), &opts.isDefault)

	return command
}

func keyListCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"ls"},
		Short:   "List keys used for signing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listKeys()
		},
	}
}

func keyDeleteCommand(opts *keyDeleteOpts) *cobra.Command {
	if opts == nil {
		opts = &keyDeleteOpts{}
	}

	command := &cobra.Command{
		Use:   "delete [flags] <key_name>...",
		Short: "Delete key from signing key list",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing key names")
			}
			opts.names = args
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteKeys(cmd.Context(), opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())

	return command
}

func addKey(ctx context.Context, opts *keyAddOpts) error {
	// set log level
	ctx = opts.LoggingFlagOpts.SetLoggerLevel(ctx)
	logger := log.GetLogger(ctx)

	signingKeys, err := configutil.LoadSigningkeysOnce()
	if err != nil {
		return err
	}
	var key config.KeySuite
	name := opts.name
	if name == "" {
		return errors.New("key name cannot be empty")
	}
	pluginName := opts.plugin
	if pluginName != "" {
		logger.Debugf("Adding key with name %v and plugin name %v", name, pluginName)
		key, err = addExternalKey(ctx, opts, pluginName, name)
		if err != nil {
			return err
		}
	} else {
		return errors.New("plugin name cannot be empty")
	}

	isDefault := opts.isDefault
	err = addKeyCore(signingKeys, key, isDefault)
	if err != nil {
		return err
	}

	if err := signingKeys.Save(); err != nil {
		return err
	}

	// write out
	logger.Debugf("Added key with name %s - {%+v}", key.Name, key.ExternalKey)
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
	mgr := plugin.NewCLIManager(dir.PluginFS())
	// Check existence of plugin with name pluginName
	_, err := mgr.Get(ctx, pluginName)
	if err != nil {
		return config.KeySuite{}, err
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

func addKeyCore(signingKeys *config.SigningKeys, key config.KeySuite, markDefault bool) error {
	if slices.Contains(signingKeys.Keys, key.Name) {
		return fmt.Errorf("signing key with name %q already exists", key.Name)
	}
	signingKeys.Keys = append(signingKeys.Keys, key)
	if markDefault {
		signingKeys.Default = key.Name
	}
	return nil
}

func updateKey(ctx context.Context, opts *keyUpdateOpts) error {
	// set log level
	ctx = opts.LoggingFlagOpts.SetLoggerLevel(ctx)
	logger := log.GetLogger(ctx)

	// initialize
	name := opts.name
	// core process
	signingKeys, err := configutil.LoadSigningkeysOnce()
	if err != nil {
		return err
	}
	if !slices.Contains(signingKeys.Keys, name) {
		return errors.New(name + ": not found")
	}
	if !opts.isDefault {
		logger.Warn("--default flag is not set, command did not take effect")
		return nil
	}
	if signingKeys.Default != name {
		signingKeys.Default = name
		if err := signingKeys.Save(); err != nil {
			return err
		}
	}

	// write out
	fmt.Printf("%s: marked as default\n", name)
	return nil
}

func listKeys() error {
	// core process
	signingKeys, err := configutil.LoadSigningkeysOnce()
	if err != nil {
		return err
	}

	// write out
	return ioutil.PrintKeyMap(os.Stdout, signingKeys.Default, signingKeys.Keys)
}

func deleteKeys(ctx context.Context, opts *keyDeleteOpts) error {
	// set log level
	ctx = opts.LoggingFlagOpts.SetLoggerLevel(ctx)
	logger := log.GetLogger(ctx)

	// core process
	signingKeys, err := configutil.LoadSigningkeysOnce()
	if err != nil {
		return err
	}

	prevDefault := signingKeys.Default
	var deletedNames []string
	for _, name := range opts.names {
		idx := slices.Index(signingKeys.Keys, name)
		if idx < 0 {
			logger.Warnf("Key %s not found, command did not take effect", name)
			return errors.New(name + ": not found")
		}
		signingKeys.Keys = slices.Delete(signingKeys.Keys, idx)
		deletedNames = append(deletedNames, name)
		if prevDefault == name {
			signingKeys.Default = ""
		}
	}
	if err := signingKeys.Save(); err != nil {
		return err
	}

	// write out
	for _, name := range deletedNames {
		if prevDefault == name {
			fmt.Printf("%s: unmarked as default\n", name)
		} else {
			fmt.Println(name)
		}
	}
	return nil
}
