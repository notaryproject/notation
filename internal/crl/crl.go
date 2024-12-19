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

package crl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	corecrl "github.com/notaryproject/notation-core-go/revocation/crl"
	"github.com/notaryproject/notation-go/log"
)

// CacheWithLog implements corecrl.Cache with logging
type CacheWithLog struct {
	corecrl.Cache

	//DiscardCacheError is set to true to enable logging the discard cache error
	//warning
	DiscardCacheError bool

	// logDiscardCrlCacheErrorOnce guarantees the discard cache error
	// warning is logged only once
	logDiscardCrlCacheErrorOnce sync.Once
}

// Get retrieves the CRL bundle with the given url
func (c *CacheWithLog) Get(ctx context.Context, url string) (*corecrl.Bundle, error) {
	if c.Cache == nil {
		return nil, errors.New("cache cannot be nil")
	}
	logger := log.GetLogger(ctx)

	bundle, err := c.Cache.Get(ctx, url)
	if err != nil && !errors.Is(err, corecrl.ErrCacheMiss) {
		if c.DiscardCacheError {
			c.logDiscardCrlCacheErrorOnce.Do(c.logDiscardCrlCacheError)
		}
		logger.Debug(err.Error())
		return nil, err
	}
	return bundle, err
}

// Set stores the CRL bundle with the given url
func (c *CacheWithLog) Set(ctx context.Context, url string, bundle *corecrl.Bundle) error {
	if c.Cache == nil {
		return errors.New("cache cannot be nil")
	}
	logger := log.GetLogger(ctx)

	err := c.Cache.Set(ctx, url, bundle)
	if err != nil {
		if c.DiscardCacheError {
			c.logDiscardCrlCacheErrorOnce.Do(c.logDiscardCrlCacheError)
		}
		logger.Debug(err.Error())
		return err
	}
	return nil
}

// logDiscardCrlCacheError logs the warning when CRL cache error is
// discarded
func (c *CacheWithLog) logDiscardCrlCacheError() {
	fmt.Fprintln(os.Stderr, "Warning: CRL cache error discarded. Enable debug log through '-d' for error details.")
}
