package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"strings"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/notation/cmd/notation/internal/cmdutil"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
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
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("missing plugin package")
			}
			return nil
		},
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
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
			return errors.New("missing plugin name")
			}
			return nil
		},
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

	// get plugin directory
	pluginDir, err := dir.PluginFS().SysPath(pluginName)
	if err != nil {
		return err
	}
	//pluginDestPath := pluginDir + "\\" + pluginBinary

	pluginExists, err := exists(pluginDir+"/"+pluginBinary)
	if err != nil {
		return err
	}

	if pluginExists {
		// if force == true, overwrite plugin
		if force {
			fmt.Printf("Overwriting plugin %s in directory %s\n", pluginBinary, pluginDir)
			if _, err := osutil.CopyToDir(pluginSrcPath,pluginDir); err != nil {
				return err
			}
			return nil
		}
		// get existing plugin metadata 
		mgr := plugin.NewCLIManager(dir.PluginFS())
		currentPlugin, err := mgr.Get(command.Context(), pluginName)

		currentPluginMetadata := &proto.GetMetadataResponse{}
		if err == nil {
			resp, err := currentPlugin.GetMetadata(command.Context(), &proto.GetMetadataRequest{})
			if err == nil {
				currentPluginMetadata = resp
			}
		}

		// Compare plugin versions
		compare := semver.Compare("v"+newPluginMetadata.Version, "v"+currentPluginMetadata.Version) 

		// copy plugin, if new plugin version is greater than current plugin version
		if compare == 1 {
			prompt := fmt.Sprintf("Are you sure you want to overwrite plugin %s_v%s with v%s?", pluginName, currentPluginMetadata.Version, newPluginMetadata.Version)
			confimred, err := cmdutil.AskForConfirmation(os.Stdin, prompt, false)
			if err != nil {
				return err
			}

			if !confimred {
				return nil
			}

			fmt.Printf("Copying plugin %s to directory %s...\n", pluginName, pluginDir)
			if _, err := osutil.CopyToDir(pluginSrcPath,pluginDir); err != nil {
				return err
			}
		}

		// do not copy plugin, if new plugin version is less than or equal to current plugin version
		if compare == -1 || compare == 0 {
			fmt.Println("Skipping plugin installation. The current version is equal to or higher than the new version.\nTo overwrite the plugin, use the --force flag.") 
		}
	}

	if !pluginExists {
		fmt.Printf("Copying plugin %s to directory %s...\n", pluginName, pluginDir)
		_, err :=osutil.CopyToDir(pluginSrcPath,pluginDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func removePlugin(command *cobra.Command, args []string) error {

	pluginName := args[0]

	// get plugin directory
	pluginDir, err := dir.PluginFS().SysPath(pluginName)
	if err != nil {
		return err
	}

	// Check if plugin directory exists 
	pluginExists, err := exists(pluginDir)
	if err != nil {
		return err
	}

	if !pluginExists {
		return errors.New("plugin does not exist")
	}

	// remove plugin directory
	return os.RemoveAll(pluginDir)
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

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
