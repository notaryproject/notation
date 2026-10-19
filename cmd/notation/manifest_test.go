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
	"errors"
	"testing"

	notationerrors "github.com/notaryproject/notation/v2/cmd/notation/internal/errors"
)

func TestParseOCILayoutReference(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		wantPath string
		wantRef  string
		wantErr  bool
	}{
		{
			name:     "tag",
			raw:      "layouts/foo:v1",
			wantPath: "layouts/foo",
			wantRef:  "v1",
		},
		{
			name:     "digest",
			raw:      "layouts/foo@" + zeroDigest,
			wantPath: "layouts/foo",
			wantRef:  zeroDigest,
		},
		{
			name:     "windows path with backslashes and a tag",
			raw:      `C:\layouts\foo:v1`,
			wantPath: `C:\layouts\foo`,
			wantRef:  "v1",
		},
		{
			name:     "windows path with forward slashes and a tag",
			raw:      "C:/layouts/foo:v1",
			wantPath: "C:/layouts/foo",
			wantRef:  "v1",
		},
		{
			name:    "windows path with backslashes and no reference",
			raw:     `C:\layouts\foo`,
			wantErr: true,
		},
		{
			name:    "windows path with forward slashes and no reference",
			raw:     "C:/layouts/foo",
			wantErr: true,
		},
		{
			name:    "no reference",
			raw:     "layouts/foo",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, ref, err := parseOCILayoutReference(tt.raw)
			if tt.wantErr {
				var missingRef notationerrors.ErrorOCILayoutMissingReference
				if !errors.As(err, &missingRef) {
					t.Fatalf("parseOCILayoutReference(%q) = (%q, %q, %v), want ErrorOCILayoutMissingReference", tt.raw, path, ref, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseOCILayoutReference(%q) returned unexpected error: %v", tt.raw, err)
			}
			if path != tt.wantPath || ref != tt.wantRef {
				t.Errorf("parseOCILayoutReference(%q) = (%q, %q), want (%q, %q)", tt.raw, path, ref, tt.wantPath, tt.wantRef)
			}
		})
	}
}
