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

// NewClient returns an *http.Client with debug log and user agent set
func NewClient(ctx context.Context, client *http.Client) *http.Client {
	if client == nil {
		client = &http.Client{}
	}
	client = trace.SetHTTPDebugLog(ctx, client)
	client.Transport = SetUserAgent(client.Transport)
	return client
}

// NewAuthClient returns an *auth.Client with debug log and user agent set
func NewAuthClient(ctx context.Context, client *http.Client) *auth.Client {
	return &auth.Client{
		Client:   NewClient(ctx, client),
		Cache:    auth.NewCache(),
		ClientID: "notation",
	}
}

type userAgentTransport struct {
	base http.RoundTripper
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := req.Clone(req.Context())
	if r.Header == nil {
		r.Header = http.Header{}
	}
	r.Header.Set("User-Agent", "notation/"+version.GetVersion())
	return t.base.RoundTrip(r)
}

// SetUserAgent sets the user agent for all out-going requests.
func SetUserAgent(rt http.RoundTripper) http.RoundTripper {
	return &userAgentTransport{
		base: rt,
	}
}
