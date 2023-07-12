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

package version

import "testing"

func TestGetVersion(t *testing.T) {
	t.Run("BuildMetadata is empty", func(t *testing.T) {
		Version = "1.0"
		BuildMetadata = ""
		v := GetVersion()
		if Version != v {
			t.Errorf("Should return Version = %s, got %s", Version, v)
		}
	})

	t.Run("BuildMetadata is not empty", func(t *testing.T) {
		Version = "1.0"
		BuildMetadata = "unreleased"
		v := GetVersion()
		want := "1.0+unreleased"
		if want != v {
			t.Errorf("Should return Version = %s, got %s", want, v)
		}
	})
}
