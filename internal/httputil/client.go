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

// NewAuthClient returns an *auth.Client
func NewAuthClient(ctx context.Context, httpClient *http.Client) *auth.Client {
	client := &auth.Client{
		Client:   httpClient,
		Cache:    auth.NewCache(),
		ClientID: "notation",
	}
	client.SetUserAgent("notation/" + version.GetVersion())
	trace.SetHTTPDebugLog(ctx, client)
	return client
}
