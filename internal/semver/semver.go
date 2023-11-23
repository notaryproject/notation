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

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

// ComparePluginVersion validates and compares two plugin semantic versions
func ComparePluginVersion(v, w string) (int, error) {
	// golang.org/x/mod/semver requires semantic version strings must begin
	// with a leading "v". Adding prefix "v" in case the input plugin version
	// does not have it.
	// Reference: https://pkg.go.dev/golang.org/x/mod/semver#pkg-overview
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	if !semver.IsValid(v) {
		return 0, fmt.Errorf("%s is not a valid semantic version", v)
	}
	if !strings.HasPrefix(w, "v") {
		w = "v" + w
	}
	if !semver.IsValid(w) {
		return 0, fmt.Errorf("%s is not a valid semantic version", w)
	}
	return semver.Compare(v, w), nil
}
