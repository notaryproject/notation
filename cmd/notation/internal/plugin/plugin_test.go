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

package plugin

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestGetPluginMetadataIfExist(t *testing.T) {
	_, err := GetPluginMetadataIfExist(context.Background(), "non-exist-plugin")
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist err, got: %v", err)
	}
}

func TestExtractPluginNameFromExecutableFileName(t *testing.T) {
	pluginName, err := ExtractPluginNameFromFileName("notation-my-plugin")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if pluginName != "my-plugin" {
		t.Fatalf("expected plugin name my-plugin, but got %s", pluginName)
	}

	pluginName, err = ExtractPluginNameFromFileName("notation-my-plugin.exe")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if pluginName != "my-plugin" {
		t.Fatalf("expected plugin name my-plugin, but got %s", pluginName)
	}

	_, err = ExtractPluginNameFromFileName("myPlugin")
	expectedErrorMsg := "invalid plugin executable file name. Plugin file name requires format notation-{plugin-name}, but got myPlugin"
	if err == nil || err.Error() != expectedErrorMsg {
		t.Fatalf("expected %s, got %v", expectedErrorMsg, err)
	}

	_, err = ExtractPluginNameFromFileName("my-plugin")
	expectedErrorMsg = "invalid plugin executable file name. Plugin file name requires format notation-{plugin-name}, but got my-plugin"
	if err == nil || err.Error() != expectedErrorMsg {
		t.Fatalf("expected %s, got %v", expectedErrorMsg, err)
	}
}
