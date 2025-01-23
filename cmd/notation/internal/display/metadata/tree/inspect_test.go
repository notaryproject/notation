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
	"os"
	"testing"
	"time"

	coresignature "github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation/internal/tree"
)

func TestAddSignedAttributes(t *testing.T) {
	t.Run("empty envelopeContent", func(t *testing.T) {
		node := tree.New("root")
		ec := &coresignature.EnvelopeContent{}
		addSignedAttributes(node, ec)
		// No error or panic expected; minimal check or just ensure it doesn't crash.
	})

	t.Run("with expiry and extented node", func(t *testing.T) {
		node := tree.New("root")
		expiryTime := time.Now().Add(time.Hour)
		ec := &coresignature.EnvelopeContent{
			Payload: coresignature.Payload{
				ContentType: "application/vnd.cncf.notary.payload.v1+json",
			},
			SignerInfo: coresignature.SignerInfo{
				SignedAttributes: coresignature.SignedAttributes{
					Expiry: expiryTime,
					ExtendedAttributes: []coresignature.Attribute{
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
		if len(signedAttrNode.Children) != 5 {
			t.Fatalf("expected 5 children, got: %v", len(signedAttrNode.Children))
		}
		// verify expiry node
		expiryNode := signedAttrNode.Children[3]
		if expiryNode.Value != fmt.Sprintf("expiry: %s", expiryTime.Format(time.ANSIC)) {
			t.Fatalf("expected expiry node, got: %v", expiryNode.Value)
		}
		// verify extended attribute node
		extendedAttrNode := signedAttrNode.Children[4]
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

func TestAddTimestamp(t *testing.T) {
	t.Run("invalid timestamp signature", func(t *testing.T) {
		node := tree.New("root")
		signerInfo := coresignature.SignerInfo{
			UnsignedAttributes: coresignature.UnsignedAttributes{
				TimestampSignature: []byte("invalid"),
			},
		}
		addTimestamp(node, signerInfo)
		if len(node.Children) == 0 {
			t.Fatal("expected node to have children")
		}
		timestampNode := node.Children[0]
		if timestampNode.Value != "timestamp signature" {
			t.Fatalf("expected 'timestamp signature' node, got %s", timestampNode.Value)
		}
		if len(timestampNode.Children) == 0 {
			t.Fatal("expected node to have children")
		}
		errNode := timestampNode.Children[0]
		expectedErrMsg := "error: failed to parse timestamp countersignature: cms: syntax error: invalid signed data: failed to convert from BER to DER: asn1: syntax error: decoding BER length octets: short form length octets value should be less or equal to the subsequent octets length"
		if errNode.Value != expectedErrMsg {
			t.Fatalf("expected error node, got %s", errNode.Value)
		}
	})

	t.Run("timestamp validation error", func(t *testing.T) {
		tsaToken, err := os.ReadFile("../testdata/TimeStampTokenWithInvalidSignature.p7s")
		if err != nil {
			t.Fatal(err)
		}
		signerInfo := coresignature.SignerInfo{
			UnsignedAttributes: coresignature.UnsignedAttributes{
				TimestampSignature: tsaToken,
			},
		}
		node := tree.New("root")
		addTimestamp(node, signerInfo)
		if len(node.Children) == 0 {
			t.Fatal("expected node to have children")
		}
		timestampNode := node.Children[0]
		if timestampNode.Value != "timestamp signature" {
			t.Fatalf("expected 'timestamp signature' node, got %s", timestampNode.Value)
		}
		if len(timestampNode.Children) == 0 {
			t.Fatal("expected node to have children")
		}
		errNode := timestampNode.Children[0]
		expectedErrMsg := "error: failed to parse timestamp countersignature: invalid TSTInfo: mismatched message"
		if errNode.Value != expectedErrMsg {
			t.Fatalf("expected error node, got %s", errNode.Value)
		}
	})
}
