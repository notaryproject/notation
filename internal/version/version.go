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

var (
	// Version shows the current notation version, optionally with pre-release.
	Version = "v1.1.0"

	// BuildMetadata stores the build metadata.
	//
	// When execute `make build` command, it will be overridden by
	// environment variable `BUILD_METADATA`. If commit tag was set,
	// BuildMetadata will be empty.
	BuildMetadata = "unreleased"

	// GitCommit stores the git HEAD commit id
	GitCommit = ""
)

// GetVersion returns the version string in SemVer 2.
func GetVersion() string {
	if BuildMetadata == "" {
		return Version
	}
	return Version + "+" + BuildMetadata
}
