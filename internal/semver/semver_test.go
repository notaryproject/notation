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

package semver

import "testing"

func TestComparePluginVersion(t *testing.T) {
	comp, err := ComparePluginVersion("v1.0.0", "v1.0.1")
	if err != nil || comp >= 0 {
		t.Fatal("expected nil err and negative comp")
	}

	comp, err = ComparePluginVersion("1.0.0", "1.0.1")
	if err != nil || comp >= 0 {
		t.Fatal("expected nil err and negative comp")
	}

	comp, err = ComparePluginVersion("1.0.1", "1.0.1")
	if err != nil || comp != 0 {
		t.Fatal("expected nil err and comp equal to 0")
	}

	expectedErrMsg := "vabc is not a valid semantic version"
	_, err = ComparePluginVersion("abc", "1.0.1")
	if err == nil || err.Error() != expectedErrMsg {
		t.Fatalf("expected err %s, got %s", expectedErrMsg, err)
	}
}
