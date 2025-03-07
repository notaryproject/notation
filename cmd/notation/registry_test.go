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

package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation/cmd/notation/internal/flag"
	"github.com/notaryproject/notation/internal/config"
)

const (
	zeroDigest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
)

func TestRegistry_getRemoteRepositoryWithReferrersAPISupported(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/test/v1/referrers/"+zeroDigest {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{ "test": "TEST" }`))
			return
		}
		t.Errorf("unexpected access: %s %q", r.Method, r.URL)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()
	uri, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("invalid test http server: %v", err)
	}
	secureOpts := flag.SecureFlagOpts{
		InsecureRegistry: true,
	}
	_, err = getRemoteRepository(context.Background(), &secureOpts, uri.Host+"/test:v1", true)
	if err != nil {
		t.Errorf("getRemoteRepository() expected nil error, but got error: %v", err)
	}
}

func TestRegistry_getRemoteRepositoryWithReferrersAPINotSupported(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/test/v1/referrers/"+zeroDigest {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		t.Errorf("unexpected access: %s %q", r.Method, r.URL)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()
	uri, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("invalid test http server: %v", err)
	}
	secureOpts := flag.SecureFlagOpts{
		InsecureRegistry: true,
	}
	_, err = getRemoteRepository(context.Background(), &secureOpts, uri.Host+"/test:v1", true)
	if err != nil {
		t.Errorf("getRemoteRepository() expected nil error, but got error: %v", err)
	}
}

func TestRegistry_getRemoteRepositoryWithReferrersTagSchema(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/test/v1/referrers/"+zeroDigest {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{ "test": "TEST" }`))
			return
		}
		t.Errorf("unexpected access: %s %q", r.Method, r.URL)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()
	uri, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("invalid test http server: %v", err)
	}
	secureOpts := flag.SecureFlagOpts{
		InsecureRegistry: true,
	}
	_, err = getRemoteRepository(context.Background(), &secureOpts, uri.Host+"/test:v1", false)
	if err != nil {
		t.Errorf("getRemoteRepository() expected nil error, but got error: %v", err)
	}
}

func TestIsRegistryInsecure(t *testing.T) {
	// for restore dir
	defer func(oldDir string) {
		dir.UserConfigDir = oldDir
		config.LoadConfigOnce = sync.OnceValues(config.LoadConfig)
	}(dir.UserConfigDir)
	// update config dir
	dir.UserConfigDir = "./internal/testdata"

	type args struct {
		target string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "hit registry", args: args{target: "reg1.io"}, want: true},
		{name: "miss registry", args: args{target: "reg2.io"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRegistryInsecure(tt.args.target); got != tt.want {
				t.Errorf("IsRegistryInsecure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRegistryInsecureMissingConfig(t *testing.T) {
	// for restore dir
	defer func(oldDir string) {
		dir.UserConfigDir = oldDir
		config.LoadConfigOnce = sync.OnceValues(config.LoadConfig)
	}(dir.UserConfigDir)
	// update config dir
	dir.UserConfigDir = "./internal/testdata2"

	type args struct {
		target string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "missing config", args: args{target: "reg1.io"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRegistryInsecure(tt.args.target); got != tt.want {
				t.Errorf("IsRegistryInsecure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRegistryInsecureConfigPermissionError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on Windows")
	}
	configDir := "./internal/testdata"
	// for restore dir
	defer func(oldDir string) error {
		// restore permission
		dir.UserConfigDir = oldDir
		config.LoadConfigOnce = sync.OnceValues(config.LoadConfig)
		return os.Chmod(filepath.Join(configDir, "config.json"), 0644)
	}(dir.UserConfigDir)

	// update config dir
	dir.UserConfigDir = configDir

	// forbid reading the file
	if err := os.Chmod(filepath.Join(configDir, "config.json"), 0000); err != nil {
		t.Error(err)
	}

	if isRegistryInsecure("reg1.io") {
		t.Error("should false because of missing config.json read permission.")
	}
}
