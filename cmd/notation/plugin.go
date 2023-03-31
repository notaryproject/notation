package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"github.com/Masterminds/semver"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/plugin/proto"
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
			notation plugin install <plugin package>
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return installPlugin(cmd, args, force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing plugin files without prompting")

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

	pluginPath := args[0]

	var pluginBinary string
	// if plugin contains a file path split and select the last element in the array
	if strings.Contains(pluginPath, "/") {
		pluginBinary = strings.Split(pluginPath, "/")[len(strings.Split(pluginPath, "/"))-1]
	} else {
		pluginBinary = pluginPath
	}

	// get plugin metadata
	cmd := exec.Command("./"+pluginPath, "get-plugin-metadata")

	output, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	var newPlugin map[string]interface{}
	err = json.Unmarshal([]byte(output), &newPlugin)
	if err != nil {
		return err
	}

	pluginName := newPlugin["name"].(string)
	newPluginVersion := newPlugin["version"].(string)
	newSemVersion, err := semver.NewVersion(newPluginVersion)
	if err != nil {
		return err
	}

	// get plugin directory
	pluginDir, err := dir.PluginFS().SysPath(pluginName)
	if err != nil {
		panic(err)
	}

	// Check if plugin directory exists
	_, err = os.Stat(pluginDir)

	// create the directory, if not exist
	if os.IsNotExist(err) {
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			return err
		}
	}

	// Check if plugin exists
	_, err = os.Stat(pluginDir + "/" + pluginBinary)

	// copy plugin, if not exist
	if os.IsNotExist(err) {
		copyPlugin(pluginPath, pluginDir+"/"+pluginBinary)
	}

	// overwrite plugin, if force flag is set
	if err == nil && force {
		fmt.Printf("Overwriting plugin %s in directory %s\n", pluginBinary, pluginDir)
		copyPlugin(pluginPath, pluginDir+"/"+pluginBinary)
	}

	// if plugin exists and force flag is not set, get plugin metadata
	if err == nil && !force {
		cmd := exec.Command(pluginDir+"/"+pluginBinary, "get-plugin-metadata")

		output, err := cmd.Output()
		if err != nil {
			return err
		}

		var currentPlugin map[string]interface{}
		err = json.Unmarshal([]byte(output), &currentPlugin)
		if err != nil {
			return err
		}

		// check if new plugin version is greater than current plugin version
		currentPluginVersion := currentPlugin["version"].(string)
		currentVersion, err := semver.NewVersion(currentPluginVersion)
		if err != nil {
			return err
		}

		// copy plugin, if new plugin version is greater than current plugin version
		if newSemVersion.GreaterThan(currentVersion) {
			var confirm string

			fmt.Printf("Detected new version %s. Current version is %s.\nDo you want to overwrite the plugin %s? [y/N]: ", newSemVersion.String(), currentVersion.String(), pluginName)
			fmt.Scanln(&confirm)

			if strings.ToLower(confirm) != "y" {
				fmt.Println("Operation cancelled.")
				return nil
			}

			fmt.Printf("Copying plugin %s to directory %s...\n", pluginName, pluginDir)
			copyPlugin(pluginPath, pluginDir+"/"+pluginBinary)
		}

		// do not copy plugin, if new plugin version is less than or equal to current plugin version
		if newSemVersion.LessThan(currentVersion) || newSemVersion.Equal(currentVersion) {
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
	err = os.RemoveAll(pluginDir)
	if err != nil {
		return err
	}

	return nil
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
