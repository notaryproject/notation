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

package revocation

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/notaryproject/notation-core-go/revocation"
	corecrl "github.com/notaryproject/notation-core-go/revocation/crl"
	"github.com/notaryproject/notation-core-go/revocation/purpose"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verifier/crl"
	clicrl "github.com/notaryproject/notation/internal/crl"
	"github.com/notaryproject/notation/internal/httputil"
)

// NewRevocationValidator returns a revocation.Validator given the certificate
// purpose
func NewRevocationValidator(ctx context.Context, purpose purpose.Purpose) (revocation.Validator, error) {
	ocspHttpClient := httputil.NewClient(ctx, &http.Client{Timeout: 2 * time.Second})
	crlFetcher, err := corecrl.NewHTTPFetcher(httputil.NewClient(ctx, &http.Client{Timeout: 5 * time.Second}))
	if err != nil {
		return nil, err
	}
	crlFetcher.DiscardCacheError = true // discard crl cache error
	cacheRoot, err := dir.CacheFS().SysPath(dir.PathCRLCache)
	if err != nil {
		return nil, err
	}
	fileCache, err := crl.NewFileCache(cacheRoot)
	if err != nil {
		// discard NewFileCache error as cache errors are not critical
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	} else {
		crlFetcher.Cache = &clicrl.CacheWithLog{
			Cache:             fileCache,
			DiscardCacheError: crlFetcher.DiscardCacheError,
		}
	}
	return revocation.NewWithOptions(revocation.Options{
		OCSPHTTPClient:   ocspHttpClient,
		CRLFetcher:       crlFetcher,
		CertChainPurpose: purpose,
	})
}
