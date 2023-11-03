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
	"testing"

	"github.com/notaryproject/notation/cmd/notation/internal/experimental"
)

const (
	zeroDigest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
)

func TestRegistry_getRemoteRepositoryWithReferrersAPISupported(t *testing.T) {
	t.Setenv("NOTATION_EXPERIMENTAL", "1")
	if experimental.IsDisabled() {
		t.Fatal("failed to enable experimental")
	}
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
	secureOpts := SecureFlagOpts{
		InsecureRegistry: true,
	}
	_, err = getRemoteRepository(context.Background(), &secureOpts, uri.Host+"/test:v1", true)
	if err != nil {
		t.Errorf("getRemoteRepository() expected nil error, but got error: %v", err)
	}
}

func TestRegistry_getRemoteRepositoryWithReferrersAPINotSupported(t *testing.T) {
	t.Setenv("NOTATION_EXPERIMENTAL", "1")
	if experimental.IsDisabled() {
		t.Fatal("failed to enable experimental")
	}
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
	secureOpts := SecureFlagOpts{
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
	secureOpts := SecureFlagOpts{
		InsecureRegistry: true,
	}
	_, err = getRemoteRepository(context.Background(), &secureOpts, uri.Host+"/test:v1", false)
	if err != nil {
		t.Errorf("getRemoteRepository() expected nil error, but got error: %v", err)
	}
}
