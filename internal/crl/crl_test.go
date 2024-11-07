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
	"os"
	"sync"
	"testing"

	corecrl "github.com/notaryproject/notation-core-go/revocation/crl"
)

func TestGet(t *testing.T) {
	cache := &CrlCacheWithLog{
		Cache: &dummyCache{},
	}
	expectedErrMsg := "cache get failed"
	_, err := cache.Get(context.Background(), "")
	if err == nil || err.Error() != expectedErrMsg {
		t.Fatalf("expected error %q, but got %q", expectedErrMsg, err)
	}

	cache = &CrlCacheWithLog{
		Cache: &dummyCache{
			cacheMiss: true,
		},
	}
	_, err = cache.Get(context.Background(), "")
	if err == nil || !errors.Is(err, corecrl.ErrCacheMiss) {
		t.Fatalf("expected error %q, but got %q", corecrl.ErrCacheMiss, err)
	}
}

func TestSet(t *testing.T) {
	cache := &CrlCacheWithLog{
		Cache: &dummyCache{},
	}
	expectedErrMsg := "cache set failed"
	err := cache.Set(context.Background(), "", nil)
	if err == nil || err.Error() != expectedErrMsg {
		t.Fatalf("expected error %q, but got %q", expectedErrMsg, err)
	}
}

func TestLogDiscardErrorOnce(t *testing.T) {
	cache := &CrlCacheWithLog{
		Cache: &dummyCache{},
	}
	oldStderr := os.Stderr
	defer func() {
		os.Stderr = oldStderr
	}()
	testFile, err := os.CreateTemp(t.TempDir(), "testNotation")
	if err != nil {
		t.Fatal(err)
	}
	defer testFile.Close()
	os.Stderr = testFile
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.Get(context.Background(), "")
		}()
	}
	wg.Wait()

	b, err := os.ReadFile(testFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	expectedMsg := "Warning: CRL cache error discarded\n"
	if string(b) != expectedMsg {
		t.Fatalf("expected to get %q, but got %q", expectedMsg, string(b))
	}
}

type dummyCache struct {
	cacheMiss bool
}

func (d *dummyCache) Get(ctx context.Context, url string) (*corecrl.Bundle, error) {
	if d.cacheMiss {
		return nil, corecrl.ErrCacheMiss
	}
	return nil, errors.New("cache get failed")
}

func (d *dummyCache) Set(ctx context.Context, url string, bundle *corecrl.Bundle) error {
	return errors.New("cache set failed")
}
