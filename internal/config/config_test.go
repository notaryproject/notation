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

package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/notaryproject/notation-go/dir"
)

func TestLoadConfigOnce(t *testing.T) {
	defer func() {
		loadConfigOnce = sync.OnceValues(loadConfig)
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
		loadConfigOnce = sync.OnceValues(loadConfig)
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

func TestIsRegistryInsecure(t *testing.T) {
	// for restore dir
	defer func(oldDir string) {
		dir.UserConfigDir = oldDir
		loadConfigOnce = sync.OnceValues(loadConfig)
	}(dir.UserConfigDir)

	// update config dir
	dir.UserConfigDir = "./testdata"
	loadConfigOnce = sync.OnceValues(loadConfig)

	type args struct {
		target string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "hit registry", args: args{target: "reg1.io"}, want: true},
		{name: "miss registry", args: args{target: "reg2.io"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRegistryInsecure(tt.args.target); got != tt.want {
				t.Errorf("IsRegistryInsecure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRegistryInsecureMissingConfig(t *testing.T) {
	// for restore dir
	defer func(oldDir string) {
		dir.UserConfigDir = oldDir
		loadConfigOnce = sync.OnceValues(loadConfig)
	}(dir.UserConfigDir)

	// update config dir
	dir.UserConfigDir = "./testdata2"
	loadConfigOnce = sync.OnceValues(loadConfig)

	type args struct {
		target string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "missing config", args: args{target: "reg1.io"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRegistryInsecure(tt.args.target); got != tt.want {
				t.Errorf("IsRegistryInsecure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRegistryInsecureConfigPermissionError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on Windows")
	}
	configDir := "./testdata"

	// for restore dir
	defer func(oldDir string) error {
		// restore permission
		dir.UserConfigDir = oldDir
		loadConfigOnce = sync.OnceValues(loadConfig)
		return os.Chmod(filepath.Join(configDir, "config.json"), 0644)
	}(dir.UserConfigDir)

	// update config dir
	dir.UserConfigDir = configDir
	loadConfigOnce = sync.OnceValues(loadConfig)

	// forbid reading the file
	if err := os.Chmod(filepath.Join(configDir, "config.json"), 0000); err != nil {
		t.Error(err)
	}
	if IsRegistryInsecure("reg1.io") {
		t.Error("should false because of missing config.json read permission.")
	}
}
