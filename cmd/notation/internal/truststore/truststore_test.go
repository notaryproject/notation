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

package truststore

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/notaryproject/notation-go/dir"
)

func TestAddCert(t *testing.T) {
	defer func(oldDir string) {
		dir.UserConfigDir = oldDir
	}(dir.UserConfigDir)

	t.Run("empty store type", func(t *testing.T) {
		expectedErrMsg := "store type cannot be empty"
		err := AddCert("", "", "test", false)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected err: %v, but got: %v", expectedErrMsg, err)
		}
	})

	t.Run("invalid store type", func(t *testing.T) {
		expectedErrMsg := "unsupported store type: invalid"
		err := AddCert("", "invalid", "test", false)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected err: %v, but got: %v", expectedErrMsg, err)
		}
	})

	t.Run("invalid store name", func(t *testing.T) {
		expectedErrMsg := "named store name needs to follow [a-zA-Z0-9_.-]+ format"
		err := AddCert("", "ca", "test%", false)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected err: %v, but got: %v", expectedErrMsg, err)
		}
	})

	t.Run("no valid certificate in file", func(t *testing.T) {
		path := filepath.FromSlash("testdata/invalid.txt")
		expectedErrMsg := "x509: malformed certificate"
		err := AddCert(path, "ca", "test", false)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected err: %v, but got: %v", expectedErrMsg, err)
		}
	})

	t.Run("cert already exists", func(t *testing.T) {
		dir.UserConfigDir = "testdata"
		path := filepath.FromSlash("testdata/self-signed.crt")
		expectedErrMsg := "certificate already exists in the Trust Store"
		err := AddCert(path, "ca", "test", false)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected err: %v, but got: %v", expectedErrMsg, err)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		path := filepath.FromSlash("../../../../internal/testdata/Empty.txt")
		expectedErr := errors.New("no valid certificate found in the empty file")
		err := AddCert(path, "ca", "test", false)
		if err == nil || err.Error() != "no valid certificate found in the file" {
			t.Fatalf("expected err: %v, but got: %v", expectedErr, err)
		}
	})

	t.Run("failed to add cert to store", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping test on Windows")
		}

		dir.UserConfigDir = t.TempDir()
		if err := os.Chmod(dir.UserConfigDir, 0000); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(dir.UserConfigDir, 0700)

		path := filepath.FromSlash("testdata/NotationTestRoot.pem")
		expectedErrMsg := "permission denied"
		err := AddCert(path, "ca", "test", false)
		if err == nil || !strings.Contains(err.Error(), expectedErrMsg) {
			t.Fatalf("expected err: %v, but got: %v", expectedErrMsg, err)
		}
	})
}

func TestDeleteAllCerts(t *testing.T) {
	defer func(oldDir string) {
		dir.UserConfigDir = oldDir
	}(dir.UserConfigDir)

	t.Run("store does not exist", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping test on Windows")
		}

		dir.UserConfigDir = "testdata"
		expectedErrMsg := `stat testdata/truststore/x509/tsa/test: no such file or directory`
		err := DeleteAllCerts("tsa", "test", true)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected err: %v, but got: %v", expectedErrMsg, err)
		}
	})
}
