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
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/log"
	notationerrors "github.com/notaryproject/notation/cmd/notation/internal/errors"
	notationplugin "github.com/notaryproject/notation/cmd/notation/internal/plugin"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/spf13/cobra"
)

const (
	notationPluginTmpDir = "notationPluginTmpDir"
	httpsPrefix          = "https://"
)

type pluginInstallOpts struct {
	cmd.LoggingFlagOpts
	pluginSource  string
	inputCheckSum string
	force         bool
}

func pluginInstallCommand(opts *pluginInstallOpts) *cobra.Command {
	if opts == nil {
		opts = &pluginInstallOpts{}
	}
	command := &cobra.Command{
		Use:   "install [flags] <plugin_src>",
		Short: "Install plugin",
		Long: `Install a Notation plugin

Example - Install plugin from file system:
  notation plugin install wabbit-plugin-v1.0.zip

Example - Install plugin from file system with user input SHA256 checksum:
  notation plugin install wabbit-plugin-v1.0.zip --checksum abcdef 

Example - Install plugin from file system regardless if it's already installed:
  notation plugin install wabbit-plugin-v1.0.zip --force

Example - Install plugin from file system with .tar.gz:
  notation plugin install wabbit-plugin-v1.0.tar.gz

Example - Install plugin from URL, SHA256 checksum is required:
  notation plugin install --checksum abcxyz https://wabbit-networks.com/intaller/linux/amd64/wabbit-plugin-v1.0.tar.gz 
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing plugin file path or URL")
			}
			opts.pluginSource = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return installPlugin(cmd, opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	command.Flags().StringVar(&opts.inputCheckSum, "checksum", "", "if set, must match the SHA256 of the plugin source")
	command.Flags().BoolVar(&opts.force, "force", false, "force to install and overwrite the plugin")
	return command
}

func installPlugin(command *cobra.Command, opts *pluginInstallOpts) error {
	// set log level
	ctx := opts.LoggingFlagOpts.InitializeLogger(command.Context())
	// get plugin source type, support file and URL
	pluginSourceType, err := getPluginSource(opts.pluginSource)
	if err != nil {
		return err
	}
	// core process
	switch pluginSourceType {
	case notationplugin.TypeFile:
		return installFromFileSystem(ctx, opts.pluginSource, opts.inputCheckSum, opts.force)
	case notationplugin.TypeURL:
		if opts.inputCheckSum == "" {
			return errors.New("install from URL requires non-empty SHA256 checksum of the plugin source")
		}
		tmpDir, err := os.MkdirTemp("", notationPluginTmpDir)
		if err != nil {
			return fmt.Errorf("failed to create notationPluginTmpDir, %w", err)
		}
		defer os.RemoveAll(tmpDir)
		downloadPath, err := notationplugin.DownloadPluginFromURL(ctx, opts.pluginSource, tmpDir)
		if err != nil {
			return fmt.Errorf("failed to download plugin from URL %s with error: %w", opts.pluginSource, err)
		}
		return installFromFileSystem(ctx, downloadPath, opts.inputCheckSum, opts.force)
	default:
		return fmt.Errorf("failed to install the plugin, plugin source type %s is not supported", pluginSourceType)
	}
}

// installFromFileSystem install the plugin from file system
func installFromFileSystem(ctx context.Context, inputPath string, inputCheckSum string, force bool) error {
	// sanity check
	inputFileStat, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to install the plugin, %w", err)
	}
	if !inputFileStat.Mode().IsRegular() {
		return fmt.Errorf("failed to install the plugin, %s is not a regular file", inputPath)
	}
	if inputCheckSum != "" {
		if err := notationplugin.ValidateCheckSum(inputPath, inputCheckSum); err != nil {
			return fmt.Errorf("failed to install the plugin, %w", err)
		}
	}
	// install the plugin based on file type
	fileType, err := osutil.DetectFileType(inputPath)
	if err != nil {
		return fmt.Errorf("failed to install the plugin, %w", err)
	}
	switch fileType {
	case notationplugin.TypeZip:
		if err := installPluginFromZip(ctx, inputPath, force); err != nil {
			return fmt.Errorf("failed to install the plugin, %w", err)
		}
		return nil
	case notationplugin.TypeGzip:
		// when file is gzip, require to be tar
		if err := installPluginFromTarGz(ctx, inputPath, force); err != nil {
			return fmt.Errorf("failed to install the plugin, %w", err)
		}
		return nil
	default:
		return errors.New("failed to install the plugin, invalid file format. Only support .tar.gz and .zip")
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
	for _, f := range archive.File {
		fmode := f.Mode()
		// requires one and only one executable file, with name in format
		// notation-{plugin-name}, exists in the zip file
		if fmode.IsRegular() && osutil.IsOwnerExecutalbeFile(fmode) {
			fileInArchive, err := f.Open()
			if err != nil {
				return err
			}
			defer fileInArchive.Close()
			err = installPluginExecutable(ctx, f.Name, fileInArchive, fmode, force)
			if errors.As(err, &notationerrors.ErrorInvalidPluginName{}) {
				logger.Warnln(err)
				continue
			}
			return err
		}
	}
	return fmt.Errorf("no valid plugin executable file was found in %s. Plugin executable file name must in format notation-{plugin-name}", zipPath)
}

// installPluginFromTarGz extracts and untar a plugin tar.gz file, validates and
// installs the plugin
func installPluginFromTarGz(ctx context.Context, tarGzPath string, force bool) error {
	logger := log.GetLogger(ctx)
	r, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer r.Close()
	decompressedStream, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer decompressedStream.Close()
	tarReader := tar.NewReader(decompressedStream)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		fmode := header.FileInfo().Mode()
		// requires one and only one executable file, with name in format
		// notation-{plugin-name}, exists in the tar.gz file
		if fmode.IsRegular() && osutil.IsOwnerExecutalbeFile(fmode) {
			err := installPluginExecutable(ctx, header.Name, tarReader, fmode, force)
			if errors.As(err, &notationerrors.ErrorInvalidPluginName{}) {
				logger.Warnln(err)
				continue
			}
			return err
		}
	}
	return fmt.Errorf("no valid plugin executable file was found in %s. Plugin executable file name must in format notation-{plugin-name}", tarGzPath)
}

// installPluginExecutable extracts, validates, and installs a plugin from
// reader
func installPluginExecutable(ctx context.Context, fileName string, fileReader io.Reader, fmode fs.FileMode, force bool) error {
	// sanity check
	pluginName, err := notationplugin.ExtractPluginNameFromExecutableFileName(fileName)
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" && filepath.Ext(fileName) != ".exe" {
		return fmt.Errorf("on Windows, plugin executable file name %s is missing the '.exe' extension", fileName)
	}
	if runtime.GOOS != "windows" && filepath.Ext(fileName) == ".exe" {
		return fmt.Errorf("on %s, plugin executable file name %s cannot have the '.exe' extension", runtime.GOOS, fileName)
	}
	// check plugin existence
	if !force {
		existed, err := notationplugin.CheckPluginExistence(ctx, pluginName)
		if err != nil {
			return fmt.Errorf("failed to check plugin existence, %w", err)
		}
		if existed {
			return fmt.Errorf("plugin %s already installed", pluginName)
		}
	}
	// extract to tmp dir
	tmpDir, err := os.MkdirTemp("", notationPluginTmpDir)
	if err != nil {
		return fmt.Errorf("failed to create notationPluginTmpDir, %w", err)
	}
	defer os.RemoveAll(tmpDir)
	tmpFilePath := filepath.Join(tmpDir, fileName)
	pluginFile, err := os.OpenFile(tmpFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fmode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(pluginFile, fileReader); err != nil {
		return err
	}
	if err := pluginFile.Close(); err != nil {
		return err
	}
	// validate plugin metadata
	pluginVersion, err := notationplugin.ValidatePluginMetadata(ctx, pluginName, tmpFilePath)
	if err != nil {
		return err
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

// getPluginSource returns the type of plugin source
func getPluginSource(source string) (notationplugin.PluginSourceType, error) {
	source = strings.TrimSpace(source)
	// check file path
	_, fileError := os.Stat(source)
	if fileError == nil {
		return notationplugin.TypeFile, nil
	}
	// check url
	if strings.HasPrefix(source, httpsPrefix) {
		return notationplugin.TypeURL, nil
	}
	// unknown
	fmt.Fprintf(os.Stdout, "%s is not a valid file: %v\n", source, fileError)
	fmt.Fprintf(os.Stdout, "%s is not a valid HTTPS URL\n", source)
	return notationplugin.TypeUnknown, fmt.Errorf("%s is an unknown plugin source", source)
}
