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

// func TestToTreeNode(t *testing.T) {
// 	t.Run("timestamp error", func(t *testing.T) {
// 		tsaToken, err := os.ReadFile("./testdata/TimeStampTokenWithInvalidSignature.p7s")
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		envelopeContent := signature.EnvelopeContent{
// 			SignerInfo: signature.SignerInfo{
// 				UnsignedAttributes: signature.UnsignedAttributes{
// 					TimestampSignature: tsaToken,
// 				},
// 			},
// 		}
// 		sig := &model.Signature{
// 			MediaType:          "mediaType",
// 			SignatureAlgorithm: "sha256",
// 			UnsignedAttributes: getUnsignedAttributes(&envelopeContent, dummyFormatter),
// 		}

// 		node := sig.ToNode("name")

// 		if len(node.Children) != 7 {
// 			t.Fatalf("expected 7 children, but got %d", len(node.Children))
// 		}

// 		unsignedNode := node.Children[4]
// 		if len(unsignedNode.Children) != 1 {
// 			t.Fatalf("expected 1 child, but got %d", len(unsignedNode.Children))
// 		}
// 		timestampNode := unsignedNode.Children[0]
// 		if len(timestampNode.Children) != 1 {
// 			t.Fatalf("expected 1 child, but got %d", len(timestampNode.Children))
// 		}
// 		if timestampError, ok := timestampNode.Children[0].Value.(string); ok {
// 			expectedErrorMsg := "error: failed to parse timestamp countersignature: invalid TSTInfo: mismatched message"
// 			if timestampError != expectedErrorMsg {
// 				t.Fatalf("expected %s, but got %s", expectedErrorMsg, timestampError)
// 			}
// 		} else {
// 			t.Fatal("expected timestamp node")
// 		}
// 	})
// }
