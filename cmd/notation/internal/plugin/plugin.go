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
	"os"
	"strings"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/log"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/plugin/proto"
	notationerrors "github.com/notaryproject/notation/cmd/notation/internal/errors"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/internal/trace"
	"github.com/notaryproject/notation/internal/version"
	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"
	"golang.org/x/mod/semver"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// PluginSourceType is an enum for plugin source
type PluginSourceType int

const (
	// PluginSourceTypeFile means plugin source is file
	PluginSourceTypeUnknown PluginSourceType = 1 + iota

	// PluginSourceTypeFile means plugin source is file
	PluginSourceTypeFile

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

// ComparePluginVersion validates and compares two plugin semantic versions
func ComparePluginVersion(v, w string) (int, error) {
	// semantic version strings must begin with a leading "v"
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	if !semver.IsValid(v) {
		return 0, fmt.Errorf("%s is not a valid semantic version", v)
	}
	if !strings.HasPrefix(w, "v") {
		w = "v" + w
	}
	if !semver.IsValid(w) {
		return 0, fmt.Errorf("%s is not a valid semantic version", w)
	}
	return semver.Compare(v, w), nil
}

// ValidateCheckSum returns nil if SHA256 of file at path equals to checkSum.
func ValidateCheckSum(path string, checkSum string) error {
	rc, err := os.Open(path)
	if err != nil {
		return err
	}
	defer rc.Close()
	dgst, err := digest.FromReader(rc)
	if err != nil {
		return err
	}
	enc := dgst.Encoded()
	if enc != strings.ToLower(checkSum) {
		return fmt.Errorf("plugin checksum does not match user input. Expecting %s", checkSum)
	}
	return nil
}

// DownloadPluginFromURL downloads plugin file from url to a tmp directory
// it returns the tmp file path of the downloaded file
func DownloadPluginFromURL(ctx context.Context, url string, tmpFile *os.File) error {
	// Get the data
	client := getClient(ctx)
	req, err := http.NewRequest("GET", url, nil)
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
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	// Write the body to file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// ExtractPluginNameFromExecutableFileName gets plugin name from plugin
// executable file name based on spec: https://github.com/notaryproject/specifications/blob/main/specs/plugin-extensibility.md#installation
func ExtractPluginNameFromExecutableFileName(execFileName string) (string, error) {
	fileName := osutil.FileNameWithoutExtension(execFileName)
	if !strings.HasPrefix(fileName, proto.Prefix) {
		return "", notationerrors.ErrorInvalidPluginName{Msg: fmt.Sprintf("invalid plugin executable file name. file name requires format notation-{plugin-name}, but got %s", fileName)}
	}
	_, pluginName, found := strings.Cut(fileName, "-")
	if !found {
		return "", notationerrors.ErrorInvalidPluginName{Msg: fmt.Sprintf("invalid plugin executable file name. file name requires format notation-{plugin-name}, but got %s", fileName)}
	}
	return pluginName, nil
}

func setHttpDebugLog(ctx context.Context, authClient *auth.Client) {
	if logrusLog, ok := log.GetLogger(ctx).(*logrus.Logger); ok && logrusLog.Level != logrus.DebugLevel {
		return
	}
	if authClient.Client == nil {
		authClient.Client = http.DefaultClient
	}
	if authClient.Client.Transport == nil {
		authClient.Client.Transport = http.DefaultTransport
	}
	authClient.Client.Transport = trace.NewTransport(authClient.Client.Transport)
}

// getClient returns an *auth.Client
func getClient(ctx context.Context) *auth.Client {
	client := &auth.Client{
		Cache:    auth.NewCache(),
		ClientID: "notation",
	}
	client.SetUserAgent("notation/" + version.GetVersion())
	setHttpDebugLog(ctx, client)
	return client
}
