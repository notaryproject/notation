package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/notation/cmd/notation/internal/cmdutil"
	"github.com/spf13/cobra"
)

func pluginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage plugins",
	}
	cmd.AddCommand(pluginListCommand())
	cmd.AddCommand(pluginInstallCommand())
	cmd.AddCommand(pluginRemoveCommand())
	return cmd
}

func pluginListCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"ls"},
		Short:   "List installed plugins",
		Long: `List installed plugins

Example - List installed Notation plugins:
  notation plugin ls
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listPlugins(cmd)
		},
	}
}

func pluginInstallCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "install [flags] <plugin package>",
		Aliases: []string{"add"},
		Short:   "Install a plugin",
		Long: `Install a plugin

		Example - Install a Notation plugin:
			notation plugin install <path to plugin executable>
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return installPlugin(cmd, args, force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing plugin files without prompting")

	return cmd
}

func pluginRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "remove [flags] <plugin>",
		Aliases: []string{"rm", "uninstall", "delete"},
		Short:   "Remove a plugin",
		Long: `Remove a plugin

		Example - Remove a Notation plugin:
			notation plugin remove <plugin>
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return removePlugin(cmd, args)
		},
	}
}

func listPlugins(command *cobra.Command) error {
	mgr := plugin.NewCLIManager(dir.PluginFS())
	pluginNames, err := mgr.List(command.Context())
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tDESCRIPTION\tVERSION\tCAPABILITIES\tERROR\t")

	var pl plugin.Plugin
	var resp *proto.GetMetadataResponse
	for _, n := range pluginNames {
		pl, err = mgr.Get(command.Context(), n)
		metaData := &proto.GetMetadataResponse{}
		if err == nil {
			resp, err = pl.GetMetadata(command.Context(), &proto.GetMetadataRequest{})
			if err == nil {
				metaData = resp
			}
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%v\t%v\t\n",
			n, metaData.Description, metaData.Version, metaData.Capabilities, err)
	}
	return tw.Flush()
}

func installPlugin(command *cobra.Command, args []string, force bool) error {
	if len(args) != 1 {
		return errors.New("missing plugin package")
	}

	pluginSrcPath := args[0]
	pluginBinary := filepath.Base(pluginSrcPath)
	pluginName := splitPluginName(pluginBinary)

	// get plugin metadata
	pl, err := plugin.NewCLIPlugin(command.Context(), pluginName, pluginSrcPath)
	newPluginMetadata := &proto.GetMetadataResponse{}
	resp, err := pl.GetMetadata(command.Context(), &proto.GetMetadataRequest{})
	if err == nil {
		newPluginMetadata = resp
	}

	newVersion, err := semver.NewVersion(newPluginMetadata.Version)
	if err != nil {
		return err
	}

	// get plugin directory
	pluginDir, err := dir.PluginFS().SysPath(pluginName)
	if err != nil {
		return err
	}

	// Check if plugin directory exists
	_, err = os.Stat(pluginDir)

	// create the directory, if not exist
	if os.IsNotExist(err) {
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			return err
		}
	}

	pluginDestPath := pluginDir + "/" + pluginBinary

	// Check if plugin exists
	_, err = os.Stat(pluginDestPath)

	// copy plugin, if not exist
	if os.IsNotExist(err) {
		copyPlugin(pluginSrcPath, pluginDestPath)
	}

	// overwrite plugin, if force flag is set
	if err == nil && force {
		fmt.Printf("Overwriting plugin %s in directory %s\n", pluginBinary, pluginDir)
		copyPlugin(pluginSrcPath, pluginDestPath)
	}

	// if plugin exists and force flag is not set, get plugin metadata
	if err == nil && !force {
		mgr := plugin.NewCLIManager(dir.PluginFS())
		currentPlugin, err := mgr.Get(command.Context(), pluginName)

		currentPluginMetadata := &proto.GetMetadataResponse{}
		if err == nil {
			resp, err := currentPlugin.GetMetadata(command.Context(), &proto.GetMetadataRequest{})
			if err == nil {
				 currentPluginMetadata = resp
			}
		}

		// convert version to semver
		currentVersion, err := semver.NewVersion(currentPluginMetadata.Version)
		if err != nil {
			return err
		}

		// copy plugin, if new plugin version is greater than current plugin version
		if newVersion.GreaterThan(currentVersion) {
			prompt := fmt.Sprintf("Are you sure you want to overwrite plugin %s v%s with v%s?", pluginName, currentVersion.String(), newVersion.String())
			confimred, err := cmdutil.AskForConfirmation(os.Stdin, prompt, false)
			if err != nil {
				return err
			}

			if !confimred {
				return nil
			}

			fmt.Printf("Copying plugin %s to directory %s...\n", pluginName, pluginDir)
			copyPlugin(pluginSrcPath, pluginDestPath)
		}

		// do not copy plugin, if new plugin version is less than or equal to current plugin version
		if newVersion.LessThan(currentVersion) || newVersion.Equal(currentVersion) {
			fmt.Println("Current version is greater than or equal to new version. Skipping plugin installation.\nUse --force flag to overwrite the plugin.")
		}
	}

	return nil
}

func removePlugin(command *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("missing plugin name")
	}

	pluginName := args[0]

	// get plugin directory
	pluginDir, err := dir.PluginFS().SysPath(pluginName)
	if err != nil {
		return err
	}

	// Check if plugin directory exists
	_, err = os.Stat(pluginDir)
	if os.IsNotExist(err) {
		return errors.New("plugin does not exist")
	}

	// remove plugin directory
	return os.RemoveAll(pluginDir)
}

func copyPlugin(src, dest string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}

	sourceFileInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}
	fileMode := sourceFileInfo.Mode()

	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	return nil
}

func splitPluginName (p string) string {
	parts := strings.Split(p, "-")
	result := strings.Join(parts[1:3], "-")
	ext := filepath.Ext(p)

	if ext != "" {
		result = strings.TrimSuffix(result, ".exe")
	}

	return result
}
