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
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/plugin/proto"
	notationerrors "github.com/notaryproject/notation/cmd/notation/internal/errors"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/opencontainers/go-digest"
)

// PluginSourceType is an enum for plugin source
type PluginSourceType string

const (
	// TypeFile means plugin source is file
	TypeFile PluginSourceType = "file"

	// TypeURL means plugin source is URL
	TypeURL PluginSourceType = "url"

	// TypeUnknown means unknown plugin source
	TypeUnknown PluginSourceType = "unknown"
)

const (
	// TypeZip means plugin file is zip
	TypeZip = "application/zip"

	// TypeGzip means plugin file is gzip
	TypeGzip = "application/x-gzip"
)

// CheckPluginExistence returns true if a plugin already exists
func CheckPluginExistence(ctx context.Context, pluginName string) (bool, error) {
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

// ValidateCheckSum returns nil if SHA256 of file at path equals to checkSum.
func ValidateCheckSum(path string, checkSum string) error {
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
		return fmt.Errorf("plugin checksum does not match user input. Expecting %s", checkSum)
	}
	return nil
}

// ValidatePluginMetadata validates plugin metadata given plugin name and path,
// returns the plugin version on success
func ValidatePluginMetadata(ctx context.Context, pluginName, path string) (string, error) {
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

// DownloadPluginFromURL downloads plugin file from url to a tmp directory
// it returns the tmp file path of the downloaded file
func DownloadPluginFromURL(ctx context.Context, url, tmpDir string) (string, error) {
	// Create the file
	tmpFilePath := filepath.Join(tmpDir, "notationPluginTmp")
	out, err := os.OpenFile(tmpFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", err
	}
	defer out.Close()
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// Check server response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}
	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}
	return tmpFilePath, err
}

// ExtractPluginNameFromExecutableFileName gets plugin name from plugin
// executable file name based on spec: https://github.com/notaryproject/specifications/blob/main/specs/plugin-extensibility.md#installation
func ExtractPluginNameFromExecutableFileName(execFileName string) (string, error) {
	fileName := osutil.FileNameWithoutExtension(execFileName)
	_, pluginName, found := strings.Cut(fileName, "-")
	if !found || !strings.HasPrefix(fileName, proto.Prefix) {
		return "", notationerrors.ErrorInvalidPluginName{Msg: fmt.Sprintf("invalid plugin executable file name. file name requires format notation-{plugin-name}, but got %s", fileName)}
	}
	return pluginName, nil
}
