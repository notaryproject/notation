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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDownloadPluginFromURL_BadStatusRedactsCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.NotFound(w, nil)
	}))
	defer server.Close()
	pluginURL := strings.Replace(server.URL, "http://", "http://user:s3cret@", 1) + "/plugin.tar.gz"

	err := DownloadPluginFromURL(context.Background(), pluginURL, io.Discard)
	if err == nil {
		t.Fatal("DownloadPluginFromURL() expected an error, but got nil")
	}
	if !strings.Contains(err.Error(), "https response bad status: 404 Not Found") {
		t.Errorf("DownloadPluginFromURL() error = %q, want the bad status", err)
	}
	if strings.Contains(err.Error(), "s3cret") {
		t.Errorf("DownloadPluginFromURL() error contains the password: %q", err)
	}
}
