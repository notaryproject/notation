package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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
	return &cobra.Command{
		Use:     "install [flags] <plugin package>",
		Aliases: []string{"add"},
		Short:   "Install a plugin",
		Long: `Install a plugin

		Example - Install a Notation plugin:
			notation plugin install <plugin package>
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return installPlugin(cmd, args)
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

func installPlugin(command *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("missing plugin package")
	}

	plugin := args[0]

	switch {
	case strings.HasSuffix(plugin, ".zip"):

		// Open the ZIP archive
		r, err := zip.OpenReader(plugin)
		if err != nil {
			log.Fatal(err)
		}
		defer r.Close()

		// find the plugin binary in the archive
		var f *zip.File
		for _, file := range r.File {
			if strings.HasPrefix(file.Name, "notation-") {
				f = file
				break
			}
		}
		if f == nil {
			return errors.New("plugin binary not found in archive")
		}

		// Open the target file for writing
		outFile, err := os.OpenFile(f.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			log.Fatal(err)
		}
		defer outFile.Close()

		// Open the file in the archive for reading
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer rc.Close()

		// Copy the file contents to the target file
		_, err = io.Copy(outFile, rc)
		if err != nil {
			log.Fatal(err)
		}
		plugin = outFile.Name()
		outFile.Close()

	case strings.HasSuffix(plugin, ".tar.gz"):
		file, err := os.Open(plugin)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		reader, err := gzip.NewReader(file)
		if err != nil {
			log.Fatal(err)
		}
		defer reader.Close()

		// create a tar reader
		tarReader := tar.NewReader(reader)

		// iterate through the files in the archive
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			// check if the fie name is the plugin binary
			if strings.HasPrefix(header.Name, "notation-") {
				// create the output file
				outFile, err := os.OpenFile(header.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, header.FileInfo().Mode())
				if err != nil {
					log.Fatal(err)
				}
				defer outFile.Close()

				// write the contents of the file to the output file
				if _, err := io.Copy(outFile, tarReader); err != nil {
					log.Fatal(err)
				}
				plugin = outFile.Name()
				outFile.Close()
			}
		}
	}

	cmd := exec.Command("./"+plugin, "get-plugin-metadata")

	output, err := cmd.Output()
	if err != nil {
		return err
	}

	var newPlugin map[string]interface{}
	err = json.Unmarshal([]byte(output), &newPlugin)
	if err != nil {
		return err
	}

	pluginName := newPlugin["name"].(string)
	newPluginVersion := newPlugin["version"].(string)

	// get plugin directory
	pluginDir, err := dir.PluginFS().SysPath(pluginName)
	if err != nil {
		return err
	}

	// Check if plugin directory exists
	_, err = os.Stat(pluginDir + "/" + plugin)

	// create the directory, if not exist
	if os.IsNotExist(err) {
		err := os.MkdirAll(pluginDir, 0755)
		if err != nil {
			return err
		}
	}

	// if plugin dir exists, get plugin metadata
	if err == nil {
		// TODO: What if plugin dir exists but the plugin doesn't?
		cmd := exec.Command(pluginDir+"/"+plugin, "get-plugin-metadata")

		output, err := cmd.Output()
		if err != nil {
			return err
		}

		var currentPlugin map[string]interface{}
		err = json.Unmarshal([]byte(output), &currentPlugin)
		if err != nil {
			return err
		}

		newSemVersion, err := semver.NewVersion(newPluginVersion)
		if err != nil {
			return err
		}

		currentPluginVersion := currentPlugin["version"].(string)
		currentVersion, err := semver.NewVersion(currentPluginVersion)
		if err != nil {
			return err
		}

		// check if currentVersion < newVersion
		if currentVersion.Compare(newSemVersion) == -1 {
			fmt.Printf("Detected new version %s. Current version is %s.\nMoving plugin %s to directory %s\n", newSemVersion.String(), currentVersion.String(), plugin, pluginDir)

			// move the plugin binary to the plugin directory
			err = os.Rename(plugin, pluginDir+"/"+plugin)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
