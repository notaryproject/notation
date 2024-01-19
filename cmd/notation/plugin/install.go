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
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/log"
	"github.com/notaryproject/notation-go/plugin"
	notationplugin "github.com/notaryproject/notation/cmd/notation/internal/plugin"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/spf13/cobra"
)

const (
	notationPluginTmpDir          = "notation-plugin"
	notationPluginDownloadTmpFile = "notation-plugin-download"
)

type pluginInstallOpts struct {
	cmd.LoggingFlagOpts
	pluginSourceType notationplugin.PluginSourceType
	pluginSource     string
	inputChecksum    string
	isFile           bool
	isURL            bool
	force            bool
}

func installCommand(opts *pluginInstallOpts) *cobra.Command {
	if opts == nil {
		opts = &pluginInstallOpts{}
	}
	command := &cobra.Command{
		Use:     "install [flags] <--file|--url> <plugin_source>",
		Aliases: []string{"add"},
		Short:   "Install a plugin",
		Long: `Install a plugin

Example - Install plugin from file system:
  notation plugin install --file wabbit-plugin-v1.0.zip

Example - Install plugin from file system with user input SHA256 checksum:
  notation plugin install --file wabbit-plugin-v1.0.zip --sha256sum 113062a462674a0e35cb5cad75a0bb2ea16e9537025531c0fd705018fcdbc17e

Example - Install plugin from file system regardless if it's already installed:
  notation plugin install --file wabbit-plugin-v1.0.zip --force

Example - Install plugin from file system with .tar.gz:
  notation plugin install --file wabbit-plugin-v1.0.tar.gz

Example - Install plugin from file system with a single plugin executable file:
  notation plugin install --file notation-wabbit-plugin

Example - Install plugin from URL, SHA256 checksum is required:
  notation plugin install --url https://wabbit-networks.com/intaller/linux/amd64/wabbit-plugin-v1.0.tar.gz --sha256sum f8a75d9234db90069d9eb5660e5374820edf36d710bd063f4ef81e7063d3810b
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				switch {
				case opts.isFile:
					return errors.New("missing plugin file path")
				case opts.isURL:
					return errors.New("missing plugin URL")
				}
				return errors.New("missing plugin source location")
			}
			if len(args) > 1 {
				return fmt.Errorf("can only install one plugin at a time, but got %v", args)
			}
			opts.pluginSource = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch {
			case opts.isFile:
				opts.pluginSourceType = notationplugin.PluginSourceTypeFile
			case opts.isURL:
				opts.pluginSourceType = notationplugin.PluginSourceTypeURL
			}
			return install(cmd, opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	command.Flags().BoolVar(&opts.isFile, "file", false, "install plugin from a file on file system")
	command.Flags().BoolVar(&opts.isURL, "url", false, fmt.Sprintf("install plugin from an HTTPS URL. The plugin download timeout is %s", notationplugin.DownloadPluginFromURLTimeout))
	command.Flags().StringVar(&opts.inputChecksum, "sha256sum", "", "must match SHA256 of the plugin source, required when \"--url\" flag is set")
	command.Flags().BoolVar(&opts.force, "force", false, "force the installation of the plugin")
	command.MarkFlagsMutuallyExclusive("file", "url")
	command.MarkFlagsOneRequired("file", "url")
	return command
}

func install(command *cobra.Command, opts *pluginInstallOpts) error {
	// set log level
	ctx := opts.LoggingFlagOpts.InitializeLogger(command.Context())
	// core process
	switch opts.pluginSourceType {
	case notationplugin.PluginSourceTypeFile:
		if err := installPlugin(ctx, opts.pluginSource, opts.inputChecksum, opts.force); err != nil {
			return fmt.Errorf("plugin installation failed: %w", err)
		}
		return nil
	case notationplugin.PluginSourceTypeURL:
		if opts.inputChecksum == "" {
			return errors.New("installing from URL requires non-empty SHA256 checksum of the plugin source")
		}
		pluginURL, err := url.Parse(opts.pluginSource)
		if err != nil {
			return fmt.Errorf("failed to parse plugin download URL %s with error: %w", pluginURL, err)
		}
		if pluginURL.Scheme != "https" {
			return fmt.Errorf("failed to download plugin from URL: only the HTTPS scheme is supported, but got %s", pluginURL.Scheme)
		}
		tmpFile, err := os.CreateTemp("", notationPluginDownloadTmpFile)
		if err != nil {
			return fmt.Errorf("failed to create temporary file required for downloading plugin: %w", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()
		fmt.Printf("Downloading plugin from %s\n", opts.pluginSource)
		err = notationplugin.DownloadPluginFromURL(ctx, opts.pluginSource, tmpFile)
		if err != nil {
			return fmt.Errorf("failed to download plugin from URL %s with error: %w", opts.pluginSource, err)
		}
		fmt.Println("Download completed")
		if err := installPlugin(ctx, tmpFile.Name(), opts.inputChecksum, opts.force); err != nil {
			return fmt.Errorf("plugin installation failed: %w", err)
		}
		return nil
	default:
		return errors.New("plugin installation failed: unknown plugin source type")
	}
}

// installPlugin installs the plugin given plugin source path
func installPlugin(ctx context.Context, inputPath string, inputChecksum string, force bool) error {
	// sanity check
	inputFileInfo, err := os.Stat(inputPath)
	if err != nil {
		return err
	}
	if !inputFileInfo.Mode().IsRegular() {
		return fmt.Errorf("%s is not a valid file", inputPath)
	}
	// checksum check
	if inputChecksum != "" {
		if err := osutil.ValidateSHA256Sum(inputPath, inputChecksum); err != nil {
			return err
		}
	}
	// install the plugin based on file type
	fileType, err := osutil.DetectFileType(inputPath)
	if err != nil {
		return err
	}
	switch fileType {
	case notationplugin.MediaTypeZip:
		rc, err := zip.OpenReader(inputPath)
		if err != nil {
			return err
		}
		defer rc.Close()
		// check for '..' in file name to avoid zip slip vulnerability
		for _, f := range rc.File {
			if strings.Contains(f.Name, "..") {
				return fmt.Errorf("file name in zip cannot contain '..', but found %q", f.Name)
			}
		}
		return installPluginFromFS(ctx, rc, force)
	case notationplugin.MediaTypeGzip:
		// when file is gzip, required to be tar
		return installPluginFromTarGz(ctx, inputPath, force)
	default:
		// input file is not in zip or gzip, try install directly
		if inputFileInfo.Size() >= osutil.MaxFileBytes {
			return fmt.Errorf("file size reached the %d MiB size limit", osutil.MaxFileBytes/1024/1024)
		}
		installOpts := plugin.CLIInstallOptions{
			PluginPath: inputPath,
			Overwrite:  force,
		}
		return installPluginWithOptions(ctx, installOpts)
	}
}

// installPluginFromFS extracts, validates and installs the plugin files
// from a fs.FS
//
// Note: zip.ReadCloser implments fs.FS
func installPluginFromFS(ctx context.Context, pluginFS fs.FS, force bool) error {
	// set up logger
	logger := log.GetLogger(ctx)
	root := "."
	// extracting all regular files from root into tmpDir
	tmpDir, err := os.MkdirTemp("", notationPluginTmpDir)
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	var pluginFileSize int64
	if err := fs.WalkDir(pluginFS, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fName := d.Name()
		if d.IsDir() && fName != root { // skip any dir in the fs except root
			return fs.SkipDir
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		// only accept regular files
		if !info.Mode().IsRegular() {
			return nil
		}
		// check for plugin file size to avoid zip bomb vulnerability
		pluginFileSize += info.Size()
		if pluginFileSize >= osutil.MaxFileBytes {
			return fmt.Errorf("total file size reached the %d MiB size limit", osutil.MaxFileBytes/1024/1024)
		}
		logger.Debugf("Extracting file %s...", fName)
		rc, err := pluginFS.Open(path)
		if err != nil {
			return err
		}
		defer rc.Close()
		tmpFilePath := filepath.Join(tmpDir, fName)
		return osutil.CopyFromReaderToDir(rc, tmpFilePath, info.Mode())
	}); err != nil {
		return err
	}
	// install core process
	installOpts := plugin.CLIInstallOptions{
		PluginPath: tmpDir,
		Overwrite:  force,
	}
	return installPluginWithOptions(ctx, installOpts)
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
	// extracting all regular files into tmpDir
	tmpDir, err := os.MkdirTemp("", notationPluginTmpDir)
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	var pluginFileSize int64
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		// check for '..' in file name to avoid zip slip vulnerability
		if strings.Contains(header.Name, "..") {
			return fmt.Errorf("file name in tar.gz cannot contain '..', but found %q", header.Name)
		}
		// only accept regular files
		if !header.FileInfo().Mode().IsRegular() {
			continue
		}
		// check for plugin file size to avoid zip bomb vulnerability
		pluginFileSize += header.FileInfo().Size()
		if pluginFileSize >= osutil.MaxFileBytes {
			return fmt.Errorf("total file size reached the %d MiB size limit", osutil.MaxFileBytes/1024/1024)
		}
		fName := filepath.Base(header.Name)
		logger.Debugf("Extracting file %s...", fName)
		tmpFilePath := filepath.Join(tmpDir, fName)
		if err := osutil.CopyFromReaderToDir(tarReader, tmpFilePath, header.FileInfo().Mode()); err != nil {
			return err
		}
	}
	// install core process
	installOpts := plugin.CLIInstallOptions{
		PluginPath: tmpDir,
		Overwrite:  force,
	}
	return installPluginWithOptions(ctx, installOpts)
}

// installPluginWithOptions installs plugin with CLIInstallOptions
func installPluginWithOptions(ctx context.Context, opts plugin.CLIInstallOptions) error {
	mgr := plugin.NewCLIManager(dir.PluginFS())
	existingPluginMetadata, newPluginMetadata, err := mgr.Install(ctx, opts)
	if err != nil {
		var errPluginDowngrade plugin.PluginDowngradeError
		if errors.As(err, &errPluginDowngrade) {
			return fmt.Errorf("%w.\nIt is not recommended to install an older version. To force the installation, use the \"--force\" option", errPluginDowngrade)
		}

		var errExeFile *plugin.PluginExecutableFileError
		if errors.As(err, &errExeFile) {
			return fmt.Errorf("%w.\nPlease ensure that the plugin executable file is compatible with %s/%s and has appropriate permissions.", err, runtime.GOOS, runtime.GOARCH)
		}

		var errMalformedPlugin *plugin.PluginMalformedError
		if errors.As(err, &errMalformedPlugin) {
			return fmt.Errorf("%w.\nPlease ensure that the plugin executable file is intact and compatible with %s/%s. Contact the plugin publisher for further assistance.", errMalformedPlugin, runtime.GOOS, runtime.GOARCH)
		}
		return err
	}
	if existingPluginMetadata != nil {
		fmt.Printf("Successfully updated plugin %s from version %s to %s\n", newPluginMetadata.Name, existingPluginMetadata.Version, newPluginMetadata.Version)
	} else {
		fmt.Printf("Successfully installed plugin %s, version %s\n", newPluginMetadata.Name, newPluginMetadata.Version)
	}
	return nil
}
