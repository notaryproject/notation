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
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/plugin/proto"
	notationerrors "github.com/notaryproject/notation/cmd/notation/internal/errors"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/internal/trace"
	"github.com/notaryproject/notation/internal/version"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// maxPluginSourceBytes specifies the limit on how many response
// bytes are allowed in the server's response to the download from URL request
var maxPluginSourceBytes int64 = 256 * 1024 * 1024 // 256 MiB

// PluginSourceType is an enum for plugin source
type PluginSourceType int

const (
	// PluginSourceTypeFile means plugin source is file
	PluginSourceTypeFile PluginSourceType = 1 + iota

	// PluginSourceTypeURL means plugin source is URL
	PluginSourceTypeURL
)

const (
	// MediaTypeZip means plugin file is zip
	MediaTypeZip = "application/zip"

	// MediaTypeGzip means plugin file is gzip
	MediaTypeGzip = "application/x-gzip"
)

// GetPluginMetadataIfExist returns plugin's metadata if it exists in Notation
func GetPluginMetadataIfExist(ctx context.Context, pluginName string) (*proto.GetMetadataResponse, error) {
	mgr := plugin.NewCLIManager(dir.PluginFS())
	plugin, err := mgr.Get(ctx, pluginName)
	if err != nil {
		return nil, err
	}
	return plugin.GetMetadata(ctx, &proto.GetMetadataRequest{})
}

// GetPluginMetadata returns plugin's metadata given plugin path
func GetPluginMetadata(ctx context.Context, pluginName, path string) (*proto.GetMetadataResponse, error) {
	plugin, err := plugin.NewCLIPlugin(ctx, pluginName, path)
	if err != nil {
		return nil, err
	}
	return plugin.GetMetadata(ctx, &proto.GetMetadataRequest{})
}

// DownloadPluginFromURL downloads plugin file from url to a tmp directory
func DownloadPluginFromURL(ctx context.Context, pluginURL string, tmpFile io.Writer) error {
	// Get the data
	client := getClient(ctx)
	req, err := http.NewRequest("GET", pluginURL, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("https response bad status: %s", resp.Status)
	}
	// Write the body to file
	lr := &io.LimitedReader{
		R: resp.Body,
		N: maxPluginSourceBytes,
	}
	_, err = io.Copy(tmpFile, lr)
	if err != nil {
		return err
	}
	if lr.N == 0 {
		return fmt.Errorf("https response reaches the %d MiB size limit", maxPluginSourceBytes)
	}
	return nil
}

// getClient returns an *auth.Client
func getClient(ctx context.Context) *auth.Client {
	client := &auth.Client{
		Cache:    auth.NewCache(),
		ClientID: "notation",
	}
	client.SetUserAgent("notation/" + version.GetVersion())
	trace.SetHttpDebugLog(ctx, client)
	return client
}

// ExtractPluginNameFromFileName checks if fileName is a valid plugin file name
// and gets plugin name from it based on spec: https://github.com/notaryproject/specifications/blob/main/specs/plugin-extensibility.md#installation
func ExtractPluginNameFromFileName(fileName string) (string, error) {
	fname := osutil.FileNameWithoutExtension(fileName)
	pluginName, found := strings.CutPrefix(fname, proto.Prefix)
	if !found {
		return "", notationerrors.ErrorInvalidPluginFileName{Msg: fmt.Sprintf("invalid plugin executable file name. Plugin file name requires format notation-{plugin-name}, but got %s", fname)}
	}
	return pluginName, nil
}
