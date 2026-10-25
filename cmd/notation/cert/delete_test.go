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

package cert

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/notaryproject/notation-go/dir"
)

func TestCertDeleteCommand(t *testing.T) {
	opts := &certDeleteOpts{}
	cmd := certDeleteCommand(opts)
	expected := &certDeleteOpts{
		storeType:  "ca",
		namedStore: "test",
		cert:       "test.crt",
		confirmed:  true,
	}
	if err := cmd.ParseFlags([]string{
		"test.crt",
		"-t", "ca",
		"-s", "test",
		"-y"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect cert delete opts: %v, got: %v", expected, opts)
	}
}

func TestCertDeleteCommand_MissingArgs(t *testing.T) {
	cmd := certDeleteCommand(nil)
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}

// A named store of "." or ".." must not resolve to the store type directory
// or to the parent of it, which would delete other trust stores.
func TestCertDelete_DotNamedStoresDoNotTouchOtherStores(t *testing.T) {
	for _, namedStore := range []string{".", ".."} {
		t.Run(namedStore, func(t *testing.T) {
			oldConfigDir := dir.UserConfigDir
			t.Cleanup(func() { dir.UserConfigDir = oldConfigDir })
			dir.UserConfigDir = t.TempDir()

			certFiles := make([]string, 0, 2)
			for _, store := range [][2]string{{"ca", "store1"}, {"tsa", "store2"}} {
				storeDir, err := dir.ConfigFS().SysPath(dir.TrustStoreDir, "x509", store[0], store[1])
				if err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(storeDir, 0o700); err != nil {
					t.Fatal(err)
				}
				certFile := filepath.Join(storeDir, "cert.pem")
				if err := os.WriteFile(certFile, []byte("cert"), 0o600); err != nil {
					t.Fatal(err)
				}
				certFiles = append(certFiles, certFile)
			}

			err := deleteCerts(&certDeleteOpts{
				storeType:  "ca",
				namedStore: namedStore,
				all:        true,
				confirmed:  true,
			})
			if err == nil {
				t.Fatalf("deleteCerts() with the named store %q expected an error, but got nil", namedStore)
			}
			for _, certFile := range certFiles {
				if _, statErr := os.Stat(certFile); statErr != nil {
					t.Errorf("%s was removed: %v", certFile, statErr)
				}
			}
		})
	}
}
