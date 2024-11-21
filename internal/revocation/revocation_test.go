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
	"os"
	"runtime"
	"testing"

	"github.com/notaryproject/notation-core-go/revocation/purpose"
	"github.com/notaryproject/notation-go/dir"
)

func TestNewRevocationValidator(t *testing.T) {
	defer func(oldCacheDir string) {
		dir.UserCacheDir = oldCacheDir
	}(dir.UserCacheDir)

	t.Run("Success", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping test on Windows")
		}
		if _, err := NewRevocationValidator(context.Background(), purpose.Timestamping); err != nil {
			t.Fatal(err)
		}
	})

	tempRoot := t.TempDir()
	t.Run("Success but without permission to create cache directory", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping test on Windows")
		}
		dir.UserCacheDir = tempRoot
		if err := os.Chmod(tempRoot, 0); err != nil {
			t.Fatal(err)
		}
		defer func() {
			// restore permission
			if err := os.Chmod(tempRoot, 0755); err != nil {
				t.Fatalf("failed to change permission: %v", err)
			}
		}()
		if _, err := NewRevocationValidator(context.Background(), purpose.Timestamping); err != nil {
			t.Fatal(err)
		}
	})
}
