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

package httputil

import (
	"context"
	"net/http"

	"github.com/notaryproject/notation/internal/trace"
	"github.com/notaryproject/notation/internal/version"
	"oras.land/oras-go/v2/registry/remote/auth"
)

var userAgent = "notation/" + version.GetVersion()

// NewAuthClient returns an *auth.Client with debug log and user agent set
func NewAuthClient(ctx context.Context, httpClient *http.Client) *auth.Client {
	httpClient = trace.SetHTTPDebugLog(ctx, httpClient)
	client := &auth.Client{
		Client:   httpClient,
		Cache:    auth.NewCache(),
		ClientID: "notation",
	}
	client.SetUserAgent(userAgent)
	return client
}

// NewClient returns an *http.Client with debug log and user agent set
func NewClient(ctx context.Context, client *http.Client) *http.Client {
	client = trace.SetHTTPDebugLog(ctx, client)
	return SetUserAgent(client)
}

type userAgentTransport struct {
	base http.RoundTripper
}

// RoundTrip returns t.Base.RoundTrip with user agent set in the request Header
func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := req.Clone(req.Context())
	if r.Header == nil {
		r.Header = http.Header{}
	}
	r.Header.Set("User-Agent", userAgent)
	return t.base.RoundTrip(r)
}

// SetUserAgent sets the user agent for all out-going requests.
func SetUserAgent(client *http.Client) *http.Client {
	if client == nil {
		client = &http.Client{}
	}
	if client.Transport == nil {
		client.Transport = http.DefaultTransport
	}
	client.Transport = &userAgentTransport{
		base: client.Transport,
	}
	return client
}
