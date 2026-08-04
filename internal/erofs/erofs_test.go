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

package erofs

import (
	"testing"
)

func TestAlignTo512(t *testing.T) {
	// already aligned — no change
	if len(alignTo512(make([]byte, 512))) != 512 {
		t.Fatal("expected 512")
	}
	// not aligned — padded up
	if len(alignTo512(make([]byte, 1))) != 512 {
		t.Fatal("expected 512")
	}
}

func TestAppendTarData(t *testing.T) {
	got := appendTarData([]byte{1, 2}, []byte{3, 4})
	if len(got) != 4 || got[0] != 1 || got[2] != 3 {
		t.Fatal("unexpected result")
	}
}

func TestExtractRootHashFromOutput(t *testing.T) {
	output := "Root hash:\tabc123def456\n"
	hash, err := extractRootHashFromOutput(output)
	if err != nil || hash != "abc123def456" {
		t.Fatalf("got %q, err %v", hash, err)
	}

	// error on empty
	if _, err := extractRootHashFromOutput(""); err == nil {
		t.Fatal("expected error")
	}
}

func TestDefaultVeritysetupOptions(t *testing.T) {
	opts := DefaultVeritysetupOptions()
	if opts.HashAlgorithm != "sha256" || opts.DataBlockSize != 512 {
		t.Fatal("unexpected defaults")
	}
}

func TestNewConverter(t *testing.T) {
	if NewConverter("").TempDir == "" {
		t.Fatal("expected non-empty TempDir")
	}
}

func TestNewVerityCalculator(t *testing.T) {
	if NewVerityCalculator("").TempDir == "" {
		t.Fatal("expected non-empty TempDir")
	}
}
