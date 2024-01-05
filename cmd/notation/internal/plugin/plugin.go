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
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/notaryproject/notation/internal/httputil"
)

// MaxPluginSourceBytes specifies the limit on how many bytes are allowed in the
// server's response to the download from URL request.
//
// The plugin source size must be strictly less than this value.
var MaxPluginSourceBytes int64 = 256 * 1024 * 1024 // 256 MiB

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

// DownloadPluginFromURLTimeout is the timeout when downloading plugin from a
// URL
const DownloadPluginFromURLTimeout = 10 * time.Minute

// DownloadPluginFromURL downloads plugin source from url to a tmp dir on file
// system. On success, it returns the downloaded file path.
func DownloadPluginFromURL(ctx context.Context, pluginURL, tmpDir string) (string, error) {
	// Get the data
	client := httputil.NewAuthClient(ctx, &http.Client{Timeout: DownloadPluginFromURLTimeout})
	req, err := http.NewRequest(http.MethodGet, pluginURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// Check server response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s %q: https response bad status: %s", resp.Request.Method, resp.Request.URL, resp.Status)
	}
	// get the downloaded file name
	var downloadedFilename string
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		_, params, err := mime.ParseMediaType(cd)
		if err == nil { // if there's an error, use the filename in URL
			downloadedFilename = params["filename"]
		}
	}
	if downloadedFilename == "" {
		downloadedFilename = path.Base(req.URL.Path)
	}
	// Write the body to file
	tmpFilePath := filepath.Join(tmpDir, downloadedFilename)
	tmpFile, err := os.Create(tmpFilePath)
	if err != nil {
		return "", err
	}
	lr := &io.LimitedReader{
		R: resp.Body,
		N: MaxPluginSourceBytes,
	}
	_, err = io.Copy(tmpFile, lr)
	if err != nil {
		tmpFile.Close()
		return "", err
	}
	if lr.N == 0 {
		tmpFile.Close()
		return "", fmt.Errorf("%s %q: https response reached the %d MiB size limit", resp.Request.Method, resp.Request.URL, MaxPluginSourceBytes/1024/1024)
	}
	tmpFile.Close()
	return tmpFile.Name(), nil
}
