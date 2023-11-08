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
	"testing"
)

func TestCheckPluginExistence(t *testing.T) {
	exist, err := CheckPluginExistence(context.Background(), "non-exist-plugin")
	if exist || err != nil {
		t.Fatalf("expected exist to be false with nil err, got: %v, %s", exist, err)
	}
}

func TestValidateCheckSum(t *testing.T) {
	expectedErrorMsg := "plugin checksum does not match user input. Expecting abcd123"
	if err := ValidateCheckSum("./testdata/test", "abcd123"); err == nil || err.Error() != expectedErrorMsg {
		t.Fatalf("expected err %s, got %v", expectedErrorMsg, err)
	}
	if err := ValidateCheckSum("./testdata/test", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"); err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
}

func TestExtractPluginNameFromExecutableFileName(t *testing.T) {
	pluginName, err := ExtractPluginNameFromExecutableFileName("notation-my-plugin")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if pluginName != "my-plugin" {
		t.Fatalf("expected plugin name my-plugin, but got %s", pluginName)
	}

	pluginName, err = ExtractPluginNameFromExecutableFileName("notation-my-plugin.exe")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if pluginName != "my-plugin" {
		t.Fatalf("expected plugin name my-plugin, but got %s", pluginName)
	}

	_, err = ExtractPluginNameFromExecutableFileName("myPlugin")
	expectedErrorMsg := "invalid plugin executable file name. file name requires format notation-{plugin-name}, but got myPlugin"
	if err == nil || err.Error() != expectedErrorMsg {
		t.Fatalf("expected %s, got %v", expectedErrorMsg, err)
	}

	_, err = ExtractPluginNameFromExecutableFileName("my-plugin")
	expectedErrorMsg = "invalid plugin executable file name. file name requires format notation-{plugin-name}, but got my-plugin"
	if err == nil || err.Error() != expectedErrorMsg {
		t.Fatalf("expected %s, got %v", expectedErrorMsg, err)
	}
}
