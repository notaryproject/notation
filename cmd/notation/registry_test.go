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
		if r.Method == http.MethodGet && r.URL.Path == "/v2/test/referrers/"+zeroDigest {
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
	_, err = getRemoteRepository(context.Background(), &secureOpts, uri.Host+"/test", true)
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
		if r.Method == http.MethodGet && r.URL.Path == "/v2/test/referrers/"+zeroDigest {
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
	_, err = getRemoteRepository(context.Background(), &secureOpts, uri.Host+"/test", true)
	if err != nil {
		t.Errorf("getRemoteRepository() expected nil error, but got error: %v", err)
	}
}

func TestRegistry_getRemoteRepositoryWithReferrersTagSchema(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/test/referrers/"+zeroDigest {
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
	_, err = getRemoteRepository(context.Background(), &secureOpts, uri.Host+"/test", false)
	if err != nil {
		t.Errorf("getRemoteRepository() expected nil error, but got error: %v", err)
	}
}
