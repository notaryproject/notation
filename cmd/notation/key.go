package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation-go/dir"

	"github.com/notaryproject/notation-go/log"
	"github.com/notaryproject/notation/cmd/notation/cert"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/ioutil"
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

type keyRemoveOpts struct {
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

Example - Remove the key from signing key list:
  notation key remove <key_name>...
`,
	}
	command.AddCommand(keyAddCommand(nil), keyUpdateCommand(nil), keyListCommand(), keyRemoveCommand(nil))

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
	command.Flags().StringVar(&opts.plugin, "plugin", "", "signing plugin name")
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

func keyRemoveCommand(opts *keyRemoveOpts) *cobra.Command {
	if opts == nil {
		opts = &keyRemoveOpts{}
	}

	command := &cobra.Command{
		Use:   "remove [flags] <key_name>...",
		Short: "Remove key from signing key list",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing key names")
			}
			opts.names = args
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return removeKeys(cmd.Context(), opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())

	return command
}

func addKey(ctx context.Context, opts *keyAddOpts) error {
	// set log level
	ctx = opts.LoggingFlagOpts.SetLoggerLevel(ctx)

	pluginConfig, err := cmd.ParseFlagMap(opts.pluginConfig, cmd.PflagPluginConfig.Name)
	if err != nil {
		return err
	}

	// core process
	exec := func(s *config.SigningKeys) error {
		return s.AddPlugin(ctx, opts.name, opts.id, opts.plugin, pluginConfig, opts.isDefault)
	}
	if err := config.LoadExecSaveSigningKeys(exec); err != nil {
		return err
	}

	if opts.isDefault {
		fmt.Printf("%s: marked as default\n", opts.name)
	} else {
		fmt.Println(opts.name)
	}

	return nil
}

func updateKey(ctx context.Context, opts *keyUpdateOpts) error {
	// set log level
	ctx = opts.LoggingFlagOpts.SetLoggerLevel(ctx)
	logger := log.GetLogger(ctx)

	if !opts.isDefault {
		logger.Warn("--default flag is not set, command did not take effect")
		return nil
	}

	// core process
	exec := func(s *config.SigningKeys) error {
		return s.UpdateDefault(opts.name)
	}
	if err := config.LoadExecSaveSigningKeys(exec); err != nil {
		return err
	}

	// write out
	fmt.Printf("%s: marked as default\n", opts.name)
	return nil
}

func listKeys() error {
	// core process
	signingKeys, err := config.LoadSigningKeys()
	if err != nil {
		return err
	}

	// write out
	return ioutil.PrintKeyMap(os.Stdout, signingKeys.Default, signingKeys.Keys)
}

func removeKeys(ctx context.Context, opts *keyRemoveOpts) error {
	// set log level
	ctx = opts.LoggingFlagOpts.SetLoggerLevel(ctx)
	logger := log.GetLogger(ctx)

	// core process
	var removedNames []string
	var prevDefault string
	exec := func(s *config.SigningKeys) error {
		if s.Default != nil {
			prevDefault = *s.Default
		}
		var err error
		removedNames, err = s.Remove(opts.names...)
		if err != nil {
			logger.Errorf("Keys removal failed to complete with error: %v", err)
		}
		return err
	}
	if err := config.LoadExecSaveSigningKeys(exec); err != nil {
		return err
	}

	// write out
	for _, name := range removedNames {
		// TODO: this is a workaround on detecting if the removed key is the
		// key generated by `notation cert generate-test` command.
		// We need to introduce a new method `notation generate-test` with
		// `--clean` flag so that key management is decoupled with certificate.
		if name == cert.GenerateTestKeyName {
			relativeKeyPath, relativeCertPath := dir.LocalKeyPath(name)
			configFS := dir.ConfigFS()
			// can ignore errors here, because they are always nil based on
			// implementation: https://github.com/notaryproject/notation-go/blob/fb79e6df1aa8cacc8c9f4e917e36ec6dc0403f37/dir/fs.go#L23
			keyPath, _ := configFS.SysPath(relativeKeyPath)
			certPath, _ := configFS.SysPath(relativeCertPath)
			fmt.Printf("Removed test key %s from Notation signing key list. Since this is the test key created by `notation cert generate-test` command, to run `generate-test` again, you have to delete the key file from %q and delete the certificate from %q\n", name, keyPath, certPath)
		} else {
			fmt.Printf("Removed test key %s from Notation signing key list. The source key file is not deleted.\n", name)
		}
		if prevDefault == name {
			fmt.Printf("%s: unmarked as default\n", name)
		}
	}
	return nil
}
