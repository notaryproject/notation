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
		Short:   "Install plugin",
		Long: `Install a plugin

Example - Install plugin from file system:
  notation plugin install --file wabbit-plugin-v1.0.zip

Example - Install plugin from file system with user input SHA256 checksum:
  notation plugin install --file wabbit-plugin-v1.0.zip --sha256sum 113062a462674a0e35cb5cad75a0bb2ea16e9537025531c0fd705018fcdbc17e

Example - Install plugin from file system regardless if it's already installed:
  notation plugin install --file wabbit-plugin-v1.0.zip --force

Example - Install plugin from file system with .tar.gz:
  notation plugin install --file wabbit-plugin-v1.0.tar.gz

Example - Install plugin from URL, SHA256 checksum is required:
  notation plugin install --url https://wabbit-networks.com/intaller/linux/amd64/wabbit-plugin-v1.0.tar.gz --sha256sum f8a75d9234db90069d9eb5660e5374820edf36d710bd063f4ef81e7063d3810b
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				if opts.isFile {
					return errors.New("missing plugin file path")
				}
				if opts.isURL {
					return errors.New("missing plugin URL")
				}
				return errors.New("missing plugin source")
			}
			if len(args) > 1 {
				return fmt.Errorf("can only insall one plugin at a time, but got %v", args)
			}
			opts.pluginSource = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.isFile {
				opts.pluginSourceType = notationplugin.PluginSourceTypeFile
			} else if opts.isURL {
				opts.pluginSourceType = notationplugin.PluginSourceTypeURL
			}
			return install(cmd, opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	command.Flags().BoolVar(&opts.isFile, "file", false, "install plugin from a file in file system")
	command.Flags().BoolVar(&opts.isURL, "url", false, "install plugin from an HTTPS URL")
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
		return installPlugin(ctx, opts.pluginSource, opts.inputChecksum, opts.force)
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
			return fmt.Errorf("failed to create notationPluginDownloadTmpFile: %w", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()
		err = notationplugin.DownloadPluginFromURL(ctx, opts.pluginSource, tmpFile)
		if err != nil {
			return fmt.Errorf("failed to download plugin from URL %s with error: %w", opts.pluginSource, err)
		}
		return installPlugin(ctx, tmpFile.Name(), opts.inputChecksum, opts.force)
	default:
		return errors.New("plugin installation failed: unknown plugin source type")
	}
}

// installPlugin installs the plugin given plugin source path
func installPlugin(ctx context.Context, inputPath string, inputChecksum string, force bool) error {
	// sanity check
	inputFileStat, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("plugin installation failed: %w", err)
	}
	if !inputFileStat.Mode().IsRegular() {
		return fmt.Errorf("plugin installation failed: %s is not a valid file", inputPath)
	}
	// checksum check
	if inputChecksum != "" {
		if err := osutil.ValidateSHA256Sum(inputPath, inputChecksum); err != nil {
			return fmt.Errorf("plugin installation failed: %w", err)
		}
	}
	// install the plugin based on file type
	fileType, err := osutil.DetectFileType(inputPath)
	if err != nil {
		return fmt.Errorf("plugin installation failed: %w", err)
	}
	switch fileType {
	case notationplugin.MediaTypeZip:
		rc, err := zip.OpenReader(inputPath)
		if err != nil {
			return fmt.Errorf("plugin installation failed: %w", err)
		}
		defer rc.Close()
		if err := installPluginFromFS(ctx, rc, force); err != nil {
			return fmt.Errorf("plugin installation failed: %w", err)
		}
		return nil
	case notationplugin.MediaTypeGzip:
		// when file is gzip, required to be tar
		if err := installPluginFromTarGz(ctx, inputPath, force); err != nil {
			return fmt.Errorf("plugin installation failed: %w", err)
		}
		return nil
	default:
		// input file is not in zip or gzip, try install directly
		installOpts := plugin.CLIInstallOptions{
			PluginPath: inputPath,
			Overwrite:  force,
		}
		if err := installPluginWithOptions(ctx, installOpts); err != nil {
			return fmt.Errorf("plugin installation failed: %w", err)
		}
		return nil
	}
}

// installPluginFromFS extracts, validates and installs the plugin files
// from a fs.FS
//
// Note: zip.ReadCloser implments fs.FS
func installPluginFromFS(ctx context.Context, pluginFs fs.FS, force bool) error {
	// set up logger
	logger := log.GetLogger(ctx)
	root := "."
	// extracting all regular files from root into tmpDir
	tmpDir, err := os.MkdirTemp("", notationPluginTmpDir)
	if err != nil {
		return fmt.Errorf("failed to create notationPluginTmpDir: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	if err := fs.WalkDir(pluginFs, root, func(path string, d fs.DirEntry, err error) error {
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
		// it is required by github-advanced-security to check for `..` in fName
		if !info.Mode().IsRegular() || strings.Contains(fName, "..") {
			return nil
		}
		logger.Debugf("Extracting file %s...", fName)
		rc, err := pluginFs.Open(path)
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
		return fmt.Errorf("failed to create notationPluginTmpDir: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		// only accept regular files
		// it is required by github-advanced-security to check for `..` in fName
		if !header.FileInfo().Mode().IsRegular() || strings.Contains(header.Name, "..") {
			continue
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
		if errors.Is(err, &plugin.PluginDowngradeError{}) {
			return fmt.Errorf("%s. %w.\nIt is not recommended to install an older version. To force the installation, use the \"--force\" option", newPluginMetadata.Name, err)
		}
		return err
	}
	if existingPluginMetadata != nil {
		fmt.Printf("Succussefully installed plugin %s, updated the version from %s to %s\n", newPluginMetadata.Name, existingPluginMetadata.Version, newPluginMetadata.Version)
	} else {
		fmt.Printf("Succussefully installed plugin %s, version %s\n", newPluginMetadata.Name, newPluginMetadata.Version)
	}
	return nil
}
