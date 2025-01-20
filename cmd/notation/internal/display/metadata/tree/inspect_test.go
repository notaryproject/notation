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

package tree

import (
	"fmt"
	"testing"
	"time"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation/internal/tree"
)

func TestInspectSignatureNoRoot(t *testing.T) {
	h := NewInspectHandler(nil)
	err := h.InspectSignature("test-digest", "application/notation", nil)
	if err == nil || err.Error() != "artifact reference is not set" {
		t.Fatalf("expected error 'artifact reference is not set', got: %v", err)
	}
}

func TestRenderNoRoot(t *testing.T) {
	h := NewInspectHandler(nil)
	err := h.Render()
	if err == nil || err.Error() != "artifact reference is not set" {
		t.Fatalf("expected error 'artifact reference is not set', got: %v", err)
	}
}

func TestAddSignedAttributes(t *testing.T) {
	t.Run("empty envelopeContent", func(t *testing.T) {
		node := tree.New("root")
		ec := &signature.EnvelopeContent{}
		addSignedAttributes(node, ec)
		// No error or panic expected; minimal check or just ensure it doesn't crash.
	})

	t.Run("with expiry and extented node", func(t *testing.T) {
		node := tree.New("root")
		expiryTime := time.Now().Add(time.Hour)
		ec := &signature.EnvelopeContent{
			SignerInfo: signature.SignerInfo{
				SignedAttributes: signature.SignedAttributes{
					Expiry: expiryTime,
					ExtendedAttributes: []signature.Attribute{
						{
							Key:   "key",
							Value: "value",
						},
					},
				},
			},
		}
		addSignedAttributes(node, ec)
		// Verify node was added; for brevity, just check no panic
		if len(node.Children) == 0 {
			t.Fatal("expected children to be added")
		}
		signedAttrNode := node.Children[0]
		if signedAttrNode.Value != "signed attributes" {
			t.Fatalf("expected name 'signed attributes', got: %v", signedAttrNode.Value)
		}
		if len(signedAttrNode.Children) != 4 {
			t.Fatalf("expected 3 children, got: %v", len(signedAttrNode.Children))
		}
		// verify expiry node
		expiryNode := signedAttrNode.Children[2]
		if expiryNode.Value != fmt.Sprintf("expiry: %s", expiryTime.Format(time.ANSIC)) {
			t.Fatalf("expected expiry node, got: %v", expiryNode.Value)
		}
		// verify extended attribute node
		extendedAttrNode := signedAttrNode.Children[3]
		if extendedAttrNode.Value != "key: value" {
			t.Fatalf("expected extended attribute node, got: %v", extendedAttrNode.Value)
		}
	})
}

func TestAddUserDefinedAttributes(t *testing.T) {
	t.Run("empty map", func(t *testing.T) {
		node := tree.New("root")
		addUserDefinedAttributes(node, nil)
		if len(node.Children) == 0 {
			t.Fatal("expected node to have children")
		}
		udaNode := node.Children[0]
		if udaNode.Value != "user defined attributes" {
			t.Fatalf("expected 'user defined attributes' node, got %s", udaNode.Value)
		}
		if len(udaNode.Children) == 0 || udaNode.Children[0].Value != "(empty)" {
			t.Fatalf("expected '(empty)' node, got %v", udaNode.Children)
		}
	})

	t.Run("non-empty map", func(t *testing.T) {
		node := tree.New("root")
		annotations := map[string]string{"key1": "val1", "key2": "val2"}
		addUserDefinedAttributes(node, annotations)
		udaNode := node.Children[0]
		if udaNode.Value != "user defined attributes" {
			t.Fatalf("expected 'user defined attributes' node, got %s", udaNode.Value)
		}
		if len(udaNode.Children) != len(annotations) {
			t.Fatalf("expected %d children, got %d", len(annotations), len(udaNode.Children))
		}
	})
}
