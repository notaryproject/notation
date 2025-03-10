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
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestResolveReference_ErrorCases(t *testing.T) {
	t.Run("invalid reference", func(t *testing.T) {
		_, _, err := resolveReference(nil, inputTypeRegistry, "invalid-format", nil, func(s string, desc ocispec.Descriptor) {
			return
		})
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid reference") {
			t.Fatalf("Expected error containing 'invalid reference', got %q", err.Error())
		}
	})

	t.Run("get manifest error", func(t *testing.T) {
		ctx := context.Background()
		testRepo := t.TempDir()
		ref := testRepo + ":v1"
		sigRepo, err := getRepository(ctx, inputTypeOCILayout, ref, nil, false)
		if err != nil {
			t.Fatalf("Failed to get repository: %v", err)
		}
		_, _, err = resolveReference(ctx, inputTypeOCILayout, ref, sigRepo, func(s string, d ocispec.Descriptor) {
			return
		})

		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to get manifest descriptor") {
			t.Fatalf("Expected error containing 'failed to get manifest descriptor', got %q", err.Error())
		}
	})
}

func TestParseReference_ErrorCases(t *testing.T) {
	// Create temp directory for OCI layout tests
	tempDir := t.TempDir()
	validDir := filepath.Join(tempDir, "validdir")
	if err := os.MkdirAll(validDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a file to test invalid OCI layout path (not a directory)
	invalidFilePath := filepath.Join(tempDir, "notadir")
	if err := os.WriteFile(invalidFilePath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		reference  string
		inputType  inputType
		wantErrMsg string
	}{
		{
			name:       "empty reference",
			reference:  "",
			inputType:  inputTypeRegistry,
			wantErrMsg: "missing user input reference",
		},
		{
			name:       "invalid registry reference format",
			reference:  "invalid-format",
			inputType:  inputTypeRegistry,
			wantErrMsg: "invalid reference",
		},
		{
			name:       "registry reference no tag or digest",
			reference:  "example.com/repo",
			inputType:  inputTypeRegistry,
			wantErrMsg: "no tag or digest",
		},
		{
			name:       "OCI layout reference missing path",
			reference:  ":tag",
			inputType:  inputTypeOCILayout,
			wantErrMsg: "invalid reference: missing oci-layout file path",
		},
		{
			name:       "OCI layout reference missing tag",
			reference:  validDir,
			inputType:  inputTypeOCILayout,
			wantErrMsg: "invalid reference: missing tag or digest",
		},
		{
			name:       "OCI layout path does not exist",
			reference:  filepath.Join(tempDir, "nonexistent") + ":tag",
			inputType:  inputTypeOCILayout,
			wantErrMsg: "failed to resolve user input reference",
		},
		{
			name:       "OCI layout path is not a directory",
			reference:  invalidFilePath + ":tag",
			inputType:  inputTypeOCILayout,
			wantErrMsg: "input path is not a dir",
		},
		{
			name:       "unsupported input type",
			reference:  "test:tag",
			inputType:  inputType(999),
			wantErrMsg: "unsupported user inputType",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parseReference(tt.reference, tt.inputType)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Fatalf("Expected error containing %q, got %q", tt.wantErrMsg, err.Error())
			}
		})
	}
}

func TestParseReference_SuccessCases(t *testing.T) {
	// Create temp directory for OCI layout tests
	tempDir := t.TempDir()
	validDir := filepath.Join(tempDir, "validdir")
	if err := os.MkdirAll(validDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Valid SHA-256 digest (64 characters)
	validDigest := "sha256:a123456789012345678901234567890123456789012345678901234567890123"

	tests := []struct {
		name          string
		reference     string
		inputType     inputType
		wantRepoRef   string
		wantTagDigRef string
	}{
		{
			name:          "registry reference with tag",
			reference:     "example.com/repo:tag1",
			inputType:     inputTypeRegistry,
			wantRepoRef:   "example.com/repo",
			wantTagDigRef: "tag1",
		},
		{
			name:          "registry reference with digest",
			reference:     "example.com/repo@" + validDigest,
			inputType:     inputTypeRegistry,
			wantRepoRef:   "example.com/repo",
			wantTagDigRef: validDigest,
		},
		{
			name:          "OCI layout reference with tag",
			reference:     validDir + ":tag1",
			inputType:     inputTypeOCILayout,
			wantRepoRef:   validDir,
			wantTagDigRef: "tag1",
		},
		{
			name:          "OCI layout reference with digest",
			reference:     validDir + "@" + validDigest,
			inputType:     inputTypeOCILayout,
			wantRepoRef:   validDir,
			wantTagDigRef: validDigest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoRef, tagDigRef, err := parseReference(tt.reference, tt.inputType)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if repoRef != tt.wantRepoRef {
				t.Errorf("Expected repository reference %q, got %q", tt.wantRepoRef, repoRef)
			}
			if tagDigRef != tt.wantTagDigRef {
				t.Errorf("Expected tag/digest reference %q, got %q", tt.wantTagDigRef, tagDigRef)
			}
		})
	}
}
