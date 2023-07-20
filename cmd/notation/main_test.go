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

package main

import (
	"os"
	"testing"
)

func Test_UnsetEnvCredential(t *testing.T) {
	const notationUsername = "NOTATION_USERNAME"
	const notationPassword = "NOTATION_PASSWORD"
	// Set environment variables for testing
	os.Setenv(notationUsername, "testuser")
	os.Setenv(notationPassword, "testpassword")
	os.Args = []string{"notation", "version"}

	main()

	// check credentials environment variables are unset
	if os.Getenv(notationUsername) != "" {
		t.Errorf("expected %s to be unset", notationUsername)
	}

	if os.Getenv(notationPassword) != "" {
		t.Errorf("expected %s to be unset", notationPassword)
	}
}
