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

package configutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/notaryproject/notation-go/dir"
)

func TestLoadConfigOnce(t *testing.T) {
	defer func() {
		LoadConfigOnce = loadConfigOnce()
	}()
	config1, err := LoadConfigOnce()
	if err != nil {
		t.Fatal("LoadConfigOnce failed.")
	}
	config2, err := LoadConfigOnce()
	if err != nil {
		t.Fatal("LoadConfigOnce failed.")
	}
	if config1 != config2 {
		t.Fatal("LoadConfigOnce should return the same config.")
	}
}

func TestLoadConfigOnceError(t *testing.T) {
	dir.UserConfigDir = t.TempDir()
	defer func() {
		dir.UserConfigDir = ""
		LoadConfigOnce = loadConfigOnce()
	}()
	if err := os.WriteFile(filepath.Join(dir.UserConfigDir, dir.PathConfigFile), []byte("invalid json"), 0600); err != nil {
		t.Fatal("Failed to create file.")
	}

	_, err := LoadConfigOnce()
	if err == nil || !strings.Contains(err.Error(), "invalid character") {
		t.Fatal("LoadConfigOnce should fail.")
	}
	_, err2 := LoadConfigOnce()
	if err != err2 {
		t.Fatal("LoadConfigOnce should return the same error.")
	}
}
