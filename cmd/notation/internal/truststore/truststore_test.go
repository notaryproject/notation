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
	"path/filepath"
	"testing"
)

func TestEmptyCertFile(t *testing.T) {
	path := filepath.FromSlash("../../../../internal/testdata/Empty.txt")
	expectedErr := errors.New("no valid certificate found in the empty file")
	err := AddCert(path, "ca", "test", false)
	if err == nil || err.Error() != "no valid certificate found in the file" {
		t.Fatalf("expected err: %v, got: %v", expectedErr, err)
	}
}
