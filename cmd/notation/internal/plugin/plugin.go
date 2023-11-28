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

	notationauth "github.com/notaryproject/notation/internal/auth"
)

// MaxPluginSourceBytes specifies the limit on how many bytes are allowed in the
// server's response to the download from URL request.
//
// It also specifies the limit of a potentail plugin executable file in a
// .tar.gz or .zip file.
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

// DownloadPluginFromURL downloads plugin file from url to a tmp directory
func DownloadPluginFromURL(ctx context.Context, pluginURL string, tmpFile io.Writer) error {
	// Get the data
	client := notationauth.NewAuthClient(ctx)
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
		N: MaxPluginSourceBytes,
	}
	_, err = io.Copy(tmpFile, lr)
	if err != nil {
		return err
	}
	if lr.N == 0 {
		return fmt.Errorf("https response reaches the %d MiB size limit", MaxPluginSourceBytes)
	}
	return nil
}
