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

package dmverity

import (
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestCreateSignatureManifest(t *testing.T) {
	sigs := []LayerSignature{
		{LayerDigest: "sha256:abc", RootHash: "hash1", Signature: []byte("sig")},
	}
	subject := ocispec.Descriptor{Digest: "sha256:subject"}

	m, err := CreateSignatureManifest(sigs, subject)
	if err != nil {
		t.Fatal(err)
	}
	if m.SchemaVersion != 2 {
		t.Fatalf("SchemaVersion = %d, want 2", m.SchemaVersion)
	}
	if len(m.Layers) != 1 {
		t.Fatalf("Layers = %d, want 1", len(m.Layers))
	}
	if m.Layers[0].Annotations["io.cncf.notary.dmverity.layer-digest"] != "sha256:abc" {
		t.Fatal("layer digest annotation mismatch")
	}
}
