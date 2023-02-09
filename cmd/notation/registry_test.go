package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	notationerrors "github.com/notaryproject/notation/cmd/notation/internal/errors"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/errcode"
)

func TestRegistry_pingReferrersAPI_Success(t *testing.T) {
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
	repo, err := remote.NewRepository(uri.Host + "/test")
	if err != nil {
		t.Fatalf("NewRepository() error = %v", err)
	}
	repo.PlainHTTP = true
	ctx := context.Background()
	err = pingReferrersAPI(ctx, repo)
	if err != nil {
		t.Errorf("pingReferrersAPI() expected nil error, but got error: %v", err)
	}
}

func TestRegistry_pingReferrersAPI_ReferrersAPINotSupported(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/test/referrers/"+zeroDigest {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{ "errorresponse": { "method": "GET", "statuscode": 404 } }`))
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
	ctx := context.Background()
	repo, err := remote.NewRepository(uri.Host + "/test")
	if err != nil {
		t.Fatalf("NewRepository() error = %v", err)
	}
	repo.PlainHTTP = true
	err = pingReferrersAPI(ctx, repo)
	var errorReferrersAPINotSupported notationerrors.ErrorReferrersAPINotSupported
	if err == nil || !errors.As(err, &errorReferrersAPINotSupported) {
		t.Errorf("pingReferrersAPI() expected ErrorReferrersAPINotSupported, but got: %v", err)
	}
}

func TestRegistry_pingReferrersAPI_Failed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/test/referrers/"+zeroDigest {
			w.WriteHeader(http.StatusOK)
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
	ctx := context.Background()
	repo, err := remote.NewRepository(uri.Host + "/test")
	if err != nil {
		t.Fatalf("NewRepository() error = %v", err)
	}
	repo.PlainHTTP = true
	err = pingReferrersAPI(ctx, repo)
	if err == nil {
		t.Errorf("pingReferrersAPI expected to get error but got nil")
	}
}

func TestRegistry_pingReferrersAPI_RepositoryNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v2/test/referrers/"+zeroDigest {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{ "errors": [ { "code": "NAME_UNKNOWN", "message": "repository name not known to registry" } ] }`))
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
	ctx := context.Background()
	expectedErr := errcode.Error{
		Code:    errcode.ErrorCodeNameUnknown,
		Message: "repository name not known to registry",
	}

	repo, err := remote.NewRepository(uri.Host + "/test")
	if err != nil {
		t.Fatalf("NewRepository() error = %v", err)
	}
	repo.PlainHTTP = true
	err = pingReferrersAPI(ctx, repo)
	if err == nil {
		t.Fatalf("pingReferrersAPI() expected error but got nil")
	}
	var ec errcode.Error
	if !errors.As(err, &ec) {
		t.Errorf("pingReferrersAPI() expected errcode.Error")
	}
	if !reflect.DeepEqual(ec, expectedErr) {
		t.Errorf("pingReferrersAPI() expected error: %v, but got: %v", expectedErr, err)
	}
}
