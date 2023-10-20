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
	"os"
	"path/filepath"
	"strings"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
)

const (
	TypeZip  = "application/zip"
	TypeGzip = "application/x-gzip"
)

type pluginInstallOpts struct {
	inputPath     string
	inputCheckSum string
	forced        bool
}

func pluginInstallCommand(opts *pluginInstallOpts) *cobra.Command {
	if opts == nil {
		opts = &pluginInstallOpts{}
	}
	command := &cobra.Command{
		Use:   "install [flags] <plugin_path>",
		Short: "Install plugin",
		Long: `Install a Notation plugin

Example - Install plugin from file system:
  notation plugin install myPlugin.zip
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return installPlugin(cmd, opts)
		},
	}
	command.Flags().StringVar(&opts.inputPath, "file", "", "file path of the plugin to be installed, only supports tar.gz or zip format")
	command.Flags().StringVar(&opts.inputCheckSum, "checksum", "", "if set, must match the SHA256 of the plugin tar.gz/zip to be installed")
	command.Flags().BoolVar(&opts.forced, "forced", false, "do not force to install and overwrite the plugin")
	command.MarkFlagRequired("file")
	return command
}

func installPlugin(command *cobra.Command, opts *pluginInstallOpts) error {
	inputPath := opts.inputPath
	// sanity check
	iputFileStat, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to install the plugin, %w", err)
	}
	if !iputFileStat.Mode().IsRegular() {
		return fmt.Errorf("failed to install the plugin, %s is not a regular file", inputPath)
	}
	// checkSum check
	if opts.inputCheckSum != "" {
		if err := validateCheckSum(inputPath, opts.inputCheckSum); err != nil {
			return fmt.Errorf("failed to install the plugin, %w", err)
		}
	}
	// install the plugin based on file type
	fileType, err := osutil.DetectFileType(inputPath)
	if err != nil {
		return fmt.Errorf("failed to install the plugin, %w", err)
	}
	switch fileType {
	case TypeZip:
		if err := installPluginFromZip(command.Context(), inputPath, opts.forced); err != nil {
			return fmt.Errorf("failed to install the plugin, %w", err)
		}
	case TypeGzip:
		if err := installPluginFromTarGz(command.Context(), inputPath, opts.forced); err != nil {
			return fmt.Errorf("failed to install the plugin, %w", err)
		}
	default:
		return errors.New("failed to install the plugin, invalid file type. Only support tar.gz and zip")
	}
	return nil
}

// validateCheckSum returns nil if SHA256 of file at path equals to checkSum.
func validateCheckSum(path string, checkSum string) error {
	r, err := os.Open(path)
	if err != nil {
		return err
	}
	defer r.Close()
	dgst, err := digest.FromReader(r)
	if err != nil {
		return err
	}
	if dgst.Encoded() != checkSum {
		return errors.New("plugin checkSum does not match user input")
	}
	return nil
}

// installPluginFromZip extracts a plugin zip file, validates and
// installs the plugin
func installPluginFromZip(ctx context.Context, zipPath string, forced bool) error {
	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer archive.Close()
	tmpDir, err := os.MkdirTemp(".", "unzipTmpDir")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	for _, f := range archive.File {
		fileMode := f.Mode()
		// only consider regular executable files in the zip
		if fileMode.IsRegular() && osutil.IsOwnerExecutalbeFile(fileMode) {
			pluginName, err := extractPluginNameFromExecutableFileName(f.Name)
			// if error is nil, we find the plugin executable file
			if err == nil {
				// check plugin existence
				if !forced {
					existed, err := checkPluginExistence(ctx, pluginName)
					if err != nil {
						return fmt.Errorf("failed to check plugin existence, %w", err)
					}
					if existed {
						return fmt.Errorf("plugin %s already existed", pluginName)
					}
				}
				// extract to tmp dir
				tmpFilePath := filepath.Join(tmpDir, filepath.Base(f.Name))
				pluginFile, err := os.OpenFile(tmpFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
				if err != nil {
					return err
				}
				fileInArchive, err := f.Open()
				if err != nil {
					return err
				}
				defer fileInArchive.Close()
				if _, err := io.Copy(pluginFile, fileInArchive); err != nil {
					return err
				}
				if err := pluginFile.Close(); err != nil {
					return err
				}
				// validate plugin metadata
				if err := validatePluginMetadata(ctx, pluginName, tmpFilePath); err != nil {
					return err
				}
				// install plugin
				pluginPath, err := dir.PluginFS().SysPath(pluginName)
				if err != nil {
					return nil
				}
				_, err = osutil.CopyToDir(tmpFilePath, pluginPath)
				if err != nil {
					return err
				}
				fmt.Printf("Succussefully installed plugin %s\n", pluginName)
				return nil
			}
		}
	}
	return errors.New("plugin executable file not found in zip")
}

// installPluginFromTarGz extracts and untar a plugin tar.gz file, validates and
// installs the plugin
func installPluginFromTarGz(ctx context.Context, tarGzPath string, forced bool) error {
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
	tmpDir, err := os.MkdirTemp(".", "untarGzTmpDir")
	if err != nil {
		return err
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
		fileMode := header.FileInfo().Mode()
		// only consider regular executable files
		if fileMode.IsRegular() && osutil.IsOwnerExecutalbeFile(fileMode) {
			pluginName, err := extractPluginNameFromExecutableFileName(header.Name)
			// if error is nil, we find the plugin executable file
			if err == nil {
				// check plugin existence
				if !forced {
					existed, err := checkPluginExistence(ctx, pluginName)
					if err != nil {
						return fmt.Errorf("failed to check plugin existence, %w", err)
					}
					if existed {
						return fmt.Errorf("plugin %s already existed", pluginName)
					}
				}
				// extract to tmp dir
				tmpFilePath := filepath.Join(tmpDir, header.Name)
				pluginFile, err := os.OpenFile(tmpFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, header.FileInfo().Mode())
				if err != nil {
					return err
				}
				if _, err := io.Copy(pluginFile, tarReader); err != nil {
					return err
				}
				if err := pluginFile.Close(); err != nil {
					return err
				}
				// validate plugin metadata
				if err := validatePluginMetadata(ctx, pluginName, tmpFilePath); err != nil {
					return err
				}
				// install plugin
				pluginPath, err := dir.PluginFS().SysPath(pluginName)
				if err != nil {
					return nil
				}
				_, err = osutil.CopyToDir(tmpFilePath, pluginPath)
				if err != nil {
					return err
				}
				fmt.Printf("Succussefully installed plugin %s\n", pluginName)
				return nil
			}
		}
	}
	return nil
}

// extractPluginNameFromExecutableFileName gets plugin name from plugin executable
// file name based on spec: https://github.com/notaryproject/specifications/blob/main/specs/plugin-extensibility.md#installation
func extractPluginNameFromExecutableFileName(execFileName string) (string, error) {
	fileName := osutil.FileNameWithoutExtension(execFileName)
	_, pluginName, found := strings.Cut(fileName, "-")
	if !found || !strings.HasPrefix(fileName, proto.Prefix) {
		return "", fmt.Errorf("invalid plugin executable file name. file name requires format notation-{plugin-name}, got %s", fileName)
	}
	return pluginName, nil
}

// checkPluginExistence returns true if a plugin already exists
func checkPluginExistence(ctx context.Context, pluginName string) (bool, error) {
	mgr := plugin.NewCLIManager(dir.PluginFS())
	_, err := mgr.Get(ctx, pluginName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// validatePluginMetadata validates plugin metadata before installation
func validatePluginMetadata(ctx context.Context, pluginName, path string) error {
	plugin, err := plugin.NewCLIPlugin(ctx, pluginName, path)
	if err != nil {
		return err
	}
	_, err = plugin.GetMetadata(ctx, &proto.GetMetadataRequest{})
	if err != nil {
		return err
	}
	return nil
}
