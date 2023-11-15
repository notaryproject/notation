// Copyright The Notary Project Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugin

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/log"
	notationplugin "github.com/notaryproject/notation/cmd/notation/internal/plugin"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/spf13/cobra"
)

const (
	notationPluginTmpDir = "notationPluginTmpDir"
)

type pluginInstallOpts struct {
	cmd.LoggingFlagOpts
	pluginSourceType notationplugin.PluginSourceType
	pluginSource     string
	inputCheckSum    string
	isFile           bool
	isUrl            bool
	force            bool
}

func pluginInstallCommand(opts *pluginInstallOpts) *cobra.Command {
	if opts == nil {
		opts = &pluginInstallOpts{}
	}
	command := &cobra.Command{
		Use:   "install [flags] <--file|--url> <plugin_source>",
		Short: "Install plugin",
		Long: `Install a plugin

Example - Install plugin from file system:
  notation plugin install --file wabbit-plugin-v1.0.zip

Example - Install plugin from file system with user input SHA256 checksum:
  notation plugin install --file wabbit-plugin-v1.0.zip --checksum abcdef 

Example - Install plugin from file system regardless if it's already installed:
  notation plugin install --file wabbit-plugin-v1.0.zip --force

Example - Install plugin from file system with .tar.gz:
  notation plugin install --file wabbit-plugin-v1.0.tar.gz

Example - Install plugin from URL, SHA256 checksum is required:
  notation plugin install --url https://wabbit-networks.com/intaller/linux/amd64/wabbit-plugin-v1.0.tar.gz --checksum abcxyz
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing plugin file path or URL")
			}
			opts.pluginSource = args[0]
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.isFile {
				opts.pluginSourceType = notationplugin.PluginSourceTypeFile
				return nil
			}
			if opts.isUrl {
				opts.pluginSourceType = notationplugin.PluginSourceTypeURL
				return nil
			}
			return errors.New("must choose one and only one flag from [--file, --url]")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return installPlugin(cmd, opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	command.Flags().BoolVar(&opts.isFile, "file", false, "if set, install plugin from a file in file system")
	command.Flags().BoolVar(&opts.isUrl, "url", false, "if set, install plugin from a URL")
	command.Flags().StringVar(&opts.inputCheckSum, "checksum", "", "must match SHA256 of the plugin source")
	command.Flags().BoolVar(&opts.force, "force", false, "force the installation of a plugin")
	command.MarkFlagsMutuallyExclusive("file", "url")
	return command
}

func installPlugin(command *cobra.Command, opts *pluginInstallOpts) error {
	// set log level
	ctx := opts.LoggingFlagOpts.InitializeLogger(command.Context())
	// core process
	switch opts.pluginSourceType {
	case notationplugin.PluginSourceTypeFile:
		return installFromFileSystem(ctx, opts.pluginSource, opts.inputCheckSum, opts.force)
	case notationplugin.PluginSourceTypeURL:
		if opts.inputCheckSum == "" {
			return errors.New("install from URL requires non-empty SHA256 checksum of the plugin source")
		}
		url, err := url.Parse(opts.pluginSource)
		if err != nil {
			return fmt.Errorf("failed to install from URL: %v", err)
		}
		if url.Scheme != "https" {
			return fmt.Errorf("failed to install from URL: %q scheme is not HTTPS", opts.pluginSource)
		}
		tmpFile, err := os.CreateTemp(".", "notationPluginDownloadTmp")
		if err != nil {
			return err
		}
		defer os.Remove(tmpFile.Name())
		err = notationplugin.DownloadPluginFromURL(ctx, opts.pluginSource, tmpFile)
		if err != nil {
			return fmt.Errorf("failed to download plugin from URL %s with error: %w", opts.pluginSource, err)
		}
		downloadPath, err := filepath.Abs(tmpFile.Name())
		if err != nil {
			return err
		}
		return installFromFileSystem(ctx, downloadPath, opts.inputCheckSum, opts.force)
	default:
		return errors.New("failed to install the plugin: plugin source type is unknown")
	}
}

// installFromFileSystem install the plugin from file system
func installFromFileSystem(ctx context.Context, inputPath string, inputCheckSum string, force bool) error {
	// sanity check
	inputFileStat, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to install the plugin: %w", err)
	}
	if !inputFileStat.Mode().IsRegular() {
		return fmt.Errorf("failed to install the plugin: %s is not a regular file", inputPath)
	}
	// checksum check
	if inputCheckSum != "" {
		if err := notationplugin.ValidateCheckSum(inputPath, inputCheckSum); err != nil {
			return fmt.Errorf("failed to install the plugin: %w", err)
		}
	}
	// install the plugin based on file type
	fileType, err := osutil.DetectFileType(inputPath)
	if err != nil {
		return fmt.Errorf("failed to install the plugin: %w", err)
	}
	switch fileType {
	case notationplugin.MediaTypeZip:
		if err := installPluginFromZip(ctx, inputPath, force); err != nil {
			return fmt.Errorf("failed to install the plugin: %w", err)
		}
		return nil
	case notationplugin.MediaTypeGzip:
		// when file is gzip, require to be tar
		if err := installPluginFromTarGz(ctx, inputPath, force); err != nil {
			return fmt.Errorf("failed to install the plugin: %w", err)
		}
		return nil
	default:
		return errors.New("failed to install the plugin: invalid file format. Only support .tar.gz and .zip")
	}
}

// installPluginFromZip extracts a plugin zip file, validates and
// installs the plugin
func installPluginFromZip(ctx context.Context, zipPath string, force bool) error {
	logger := log.GetLogger(ctx)
	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer archive.Close()
	// require one and only one file with name in the format
	// `notation-{plugin-name}`
	for _, f := range archive.File {
		if !f.Mode().IsRegular() || strings.Contains(f.Name, "..") {
			continue
		}
		// validate and get plugin name from file name
		pluginName, err := notationplugin.ExtractPluginNameFromFileName(f.Name)
		if err != nil {
			logger.Infoln(err)
			continue
		}
		fileInArchive, err := f.Open()
		if err != nil {
			return err
		}
		defer fileInArchive.Close()
		return installPluginExecutable(ctx, f.Name, pluginName, fileInArchive, force)
	}
	return fmt.Errorf("no valid plugin file was found in %s. Plugin file name must in format notation-{plugin-name}", zipPath)
}

// installPluginFromTarGz extracts and untar a plugin tar.gz file, validates and
// installs the plugin
func installPluginFromTarGz(ctx context.Context, tarGzPath string, force bool) error {
	logger := log.GetLogger(ctx)
	rc, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer rc.Close()
	decompressedStream, err := gzip.NewReader(rc)
	if err != nil {
		return err
	}
	defer decompressedStream.Close()
	tarReader := tar.NewReader(decompressedStream)
	// require one and only one file with name in the format
	// `notation-{plugin-name}`
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if !header.FileInfo().Mode().IsRegular() || strings.Contains(header.Name, "..") {
			continue
		}
		// validate and get plugin name from file name
		pluginName, err := notationplugin.ExtractPluginNameFromFileName(header.Name)
		if err != nil {
			logger.Infoln(err)
			continue
		}
		return installPluginExecutable(ctx, header.Name, pluginName, tarReader, force)
	}
	return fmt.Errorf("no valid plugin file was found in %s. Plugin file name must in format notation-{plugin-name}", tarGzPath)
}

// installPluginExecutable extracts, validates, and installs a plugin file
func installPluginExecutable(ctx context.Context, fileName string, pluginName string, fileReader io.Reader, force bool) error {
	// sanity check
	if runtime.GOOS == "windows" && filepath.Ext(fileName) != ".exe" {
		return fmt.Errorf("on Windows, plugin executable file %s is missing the '.exe' extension", fileName)
	}
	if runtime.GOOS != "windows" && filepath.Ext(fileName) == ".exe" {
		return fmt.Errorf("on %s, plugin executable file %s cannot have the '.exe' extension", runtime.GOOS, fileName)
	}
	// extract to tmp dir
	tmpDir, err := os.MkdirTemp(".", notationPluginTmpDir)
	if err != nil {
		return fmt.Errorf("failed to create notationPluginTmpDir: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	tmpFilePath := filepath.Join(tmpDir, fileName)
	pluginFile, err := os.OpenFile(tmpFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		return err
	}
	if _, err := io.Copy(pluginFile, fileReader); err != nil {
		if err := pluginFile.Close(); err != nil {
			return err
		}
		return err
	}
	if err := pluginFile.Close(); err != nil {
		return err
	}
	// get plugin metadata
	pluginMetadata, err := notationplugin.GetPluginMetadata(ctx, pluginName, tmpFilePath)
	if err != nil {
		return err
	}
	pluginVersion := pluginMetadata.Version
	// check plugin existence and version
	if !force {
		currentPluginMetadata, err := notationplugin.GetPluginMetadataIfExist(ctx, pluginName)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err == nil { // plugin already installed
			comp, err := notationplugin.ComparePluginVersion(pluginVersion, currentPluginMetadata.Version)
			if err != nil {
				return err
			}
			if comp < 0 {
				return fmt.Errorf("%s current version %s is larger than the installing version %s", pluginName, currentPluginMetadata.Version, pluginVersion)
			}
			if comp == 0 {
				// if version is the same, no action is needed and no error is
				// returned
				fmt.Printf("%s with version %s already installed\n", pluginName, currentPluginMetadata.Version)
				return nil
			}
		}
	}
	// install plugin
	pluginPath, err := dir.PluginFS().SysPath(pluginName)
	if err != nil {
		return err
	}
	_, err = osutil.CopyToDir(tmpFilePath, pluginPath)
	if err != nil {
		return err
	}
	// plugin is always executable
	pluginFilePath := filepath.Join(pluginPath, fileName)
	err = os.Chmod(pluginFilePath, 0700)
	if err != nil {
		return err
	}
	fmt.Printf("Succussefully installed plugin %s, version %s\n", pluginName, pluginVersion)
	return nil
}
