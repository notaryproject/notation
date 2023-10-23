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
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/log"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/plugin/proto"
	notationerrors "github.com/notaryproject/notation/cmd/notation/internal/errors"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
)

const (
	TypeZip  = "application/zip"
	TypeGzip = "application/x-gzip"
)

const notationPluginTmp = "notationPluginTmp"

type pluginInstallOpts struct {
	cmd.LoggingFlagOpts
	inputPath     string
	inputURL      string
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
  notation plugin install --file myPlugin.zip --checksum abcdef

Example - Install plugin from URL:
  notation plugin install https://wabbit-networks.com/intaller/linux/amd64/wabbit-plugin-v1.0.tar.gz --checksum abcxyz
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && opts.inputPath == "" {
				return errors.New("missing plugin URL or file path")
			}
			if len(args) != 0 && opts.inputPath != "" {
				return errors.New("can install from either plugin URL or file path, got both")
			}
			if len(args) != 0 {
				opts.inputURL = args[0]
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return installPlugin(cmd, opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	command.Flags().StringVar(&opts.inputPath, "file", "", "file path of the plugin to be installed, only supports tar.gz and zip format")
	command.Flags().StringVar(&opts.inputCheckSum, "checksum", "", "if set, must match the SHA256 of the plugin tar.gz/zip to be installed")
	command.Flags().BoolVar(&opts.force, "force", false, "force to install and overwrite the plugin")
	command.MarkFlagRequired("checksum")
	return command
}

func installPlugin(command *cobra.Command, opts *pluginInstallOpts) error {
	// set log level
	ctx := opts.LoggingFlagOpts.InitializeLogger(command.Context())

	inputPath := opts.inputPath
	// install from URL
	if opts.inputURL != "" {
		inputPath = notationPluginTmp
		if err := downloadFromURL(ctx, inputPath, opts.inputURL); err != nil {
			return err
		}
		defer os.Remove(inputPath)
	}

	// sanity check
	inputFileStat, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to install the plugin, %w", err)
	}
	if !inputFileStat.Mode().IsRegular() {
		return fmt.Errorf("failed to install the plugin, %s is not a regular file", inputPath)
	}
	// checkSum check
	if err := validateCheckSum(inputPath, opts.inputCheckSum); err != nil {
		return fmt.Errorf("failed to install the plugin, %w", err)
	}
	// install the plugin based on file type
	fileType, err := osutil.DetectFileType(inputPath)
	if err != nil {
		return fmt.Errorf("failed to install the plugin, %w", err)
	}
	switch fileType {
	case TypeZip:
		if err := installPluginFromZip(ctx, inputPath, opts.force); err != nil {
			return fmt.Errorf("failed to install the plugin, %w", err)
		}
	case TypeGzip:
		if err := installPluginFromTarGz(ctx, inputPath, opts.force); err != nil {
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
	enc := dgst.Encoded()
	if enc != checkSum {
		return fmt.Errorf("plugin checkSum does not match user input. User input is %s, got %s", checkSum, enc)
	}
	return nil
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
		// only consider regular executable files
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
	return errors.New("valid plugin executable file not found in zip")
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
		// only consider regular executable files
		if fmode.IsRegular() && osutil.IsOwnerExecutalbeFile(fmode) {
			err := installPluginExecutable(ctx, header.Name, tarReader, fmode, force)
			if errors.As(err, &notationerrors.ErrorInvalidPluginName{}) {
				logger.Warnln(err)
				continue
			}
			return err
		}
	}
	return errors.New("valid plugin executable file not found in tar.gz")
}

// installPluginExecutable extracts, validates, and installs a plugin from
// reader
func installPluginExecutable(ctx context.Context, fileName string, fileReader io.Reader, fmode fs.FileMode, force bool) error {
	pluginName, err := extractPluginNameFromExecutableFileName(fileName)
	if err != nil {
		return err
	}
	// check plugin existence
	if !force {
		existed, err := checkPluginExistence(ctx, pluginName)
		if err != nil {
			return fmt.Errorf("failed to check plugin existence, %w", err)
		}
		if existed {
			return fmt.Errorf("plugin %s already installed", pluginName)
		}
	}
	// extract to tmp dir
	tmpDir, err := os.MkdirTemp(".", notationPluginTmp)
	if err != nil {
		return fmt.Errorf("failed to create notationPluginTmp, %w", err)
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
	pluginVersion, err := validatePluginMetadata(ctx, pluginName, tmpFilePath)
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
	pluginFilePath := filepath.Join(pluginPath, filepath.Base(tmpFilePath))
	err = os.Chmod(pluginFilePath, 0700)
	if err != nil {
		return err
	}

	fmt.Printf("Succussefully installed plugin %s, version %s\n", pluginName, pluginVersion)
	return nil
}

// extractPluginNameFromExecutableFileName gets plugin name from plugin
// executable file name based on spec: https://github.com/notaryproject/specifications/blob/main/specs/plugin-extensibility.md#installation
func extractPluginNameFromExecutableFileName(execFileName string) (string, error) {
	fileName := osutil.FileNameWithoutExtension(execFileName)
	_, pluginName, found := strings.Cut(fileName, "-")
	if !found || !strings.HasPrefix(fileName, proto.Prefix) {
		return "", notationerrors.ErrorInvalidPluginName{Msg: fmt.Sprintf("invalid plugin executable file name. file name requires format notation-{plugin-name}, got %s", fileName)}
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
// returns the plugin version on success
func validatePluginMetadata(ctx context.Context, pluginName, path string) (string, error) {
	plugin, err := plugin.NewCLIPlugin(ctx, pluginName, path)
	if err != nil {
		return "", err
	}
	metadata, err := plugin.GetMetadata(ctx, &proto.GetMetadataRequest{})
	if err != nil {
		return "", err
	}
	return metadata.Version, nil
}

func downloadFromURL(ctx context.Context, filePath, url string) error {
	// Create the file
	out, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
