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

package envelope

import (
	"os"
	"testing"
	"time"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation/internal/ioutil"
)

func TestGetUnsignedAttributes(t *testing.T) {
	envContent := &signature.EnvelopeContent{
		SignerInfo: signature.SignerInfo{
			UnsignedAttributes: signature.UnsignedAttributes{
				TimestampSignature: []byte("invalid"),
			},
		},
	}
	expectedErrMsg := "failed to parse timestamp countersignature: cms: syntax error: invalid signed data: failed to convert from BER to DER: asn1: syntax error: decoding BER length octets: short form length octets value should be less or equal to the subsequent octets length"
	unsignedAttr := getUnsignedAttributes(envContent)
	val, ok := unsignedAttr["timestampSignature"].(Timestamp)
	if !ok {
		t.Fatal("expected to have timestampSignature")
	}
	if val.Error != expectedErrMsg {
		t.Fatalf("expected %s, but got %s", expectedErrMsg, val.Error)
	}
}

func TestGetSignedAttributes(t *testing.T) {
	expiry := time.Now()
	envContent := &signature.EnvelopeContent{
		SignerInfo: signature.SignerInfo{
			SignedAttributes: signature.SignedAttributes{
				Expiry: expiry,
				ExtendedAttributes: []signature.Attribute{
					{
						Key:   "keyName",
						Value: "value",
					},
				},
			},
		},
	}
	signedAttr := getSignedAttributes(envContent)
	val, ok := signedAttr["expiry"].(ioutil.Time)
	if !ok {
		t.Fatal("expected to have expiry")
	}
	if val != ioutil.Time(expiry) {
		t.Fatalf("expected %v, but got %v", expiry, val)
	}

	attr, ok := signedAttr["keyName"].(string)
	if !ok {
		t.Fatal("expected to have keyName")
	}
	if attr != "value" {
		t.Fatalf("expected value, but got %s", attr)
	}
}

func TestParseTimestamp(t *testing.T) {
	t.Run("invalid timestamp signature", func(t *testing.T) {
		signerInfo := signature.SignerInfo{
			UnsignedAttributes: signature.UnsignedAttributes{
				TimestampSignature: []byte("invalid"),
			},
		}
		val := parseTimestamp(signerInfo)
		expectedErrMsg := "failed to parse timestamp countersignature: cms: syntax error: invalid signed data: failed to convert from BER to DER: asn1: syntax error: decoding BER length octets: short form length octets value should be less or equal to the subsequent octets length"
		if val.Error != expectedErrMsg {
			t.Fatalf("expected %s, but got %s", expectedErrMsg, val.Error)
		}
	})

	t.Run("timestamp validation error", func(t *testing.T) {
		tsaToken, err := os.ReadFile("./testdata/TimeStampTokenWithInvalidSignature.p7s")
		if err != nil {
			t.Fatal(err)
		}

		signerInfo := signature.SignerInfo{
			UnsignedAttributes: signature.UnsignedAttributes{
				TimestampSignature: tsaToken,
			},
		}
		val := parseTimestamp(signerInfo)
		expectedErrMsg := "failed to parse timestamp countersignature: invalid TSTInfo: mismatched message"

		if val.Error != expectedErrMsg {
			t.Fatalf("expected %s, but got %s", expectedErrMsg, val.Error)
		}
	})
}

func TestToNode(t *testing.T) {
	t.Run("timestamp error", func(t *testing.T) {
		tsaToken, err := os.ReadFile("./testdata/TimeStampTokenWithInvalidSignature.p7s")
		if err != nil {
			t.Fatal(err)
		}

		envelopeContent := signature.EnvelopeContent{
			SignerInfo: signature.SignerInfo{
				UnsignedAttributes: signature.UnsignedAttributes{
					TimestampSignature: tsaToken,
				},
			},
		}
		sig := &Signature{
			MediaType:          "mediaType",
			SignatureAlgorithm: "sha256",
			UnsignedAttributes: getUnsignedAttributes(&envelopeContent),
		}

		node := sig.ToNode("name")

		if len(node.Children) != 7 {
			t.Fatalf("expected 7 children, but got %d", len(node.Children))
		}

		unsignedNode := node.Children[4]
		if len(unsignedNode.Children) != 1 {
			t.Fatalf("expected 1 child, but got %d", len(unsignedNode.Children))
		}
		timestampNode := unsignedNode.Children[0]
		if len(timestampNode.Children) != 1 {
			t.Fatalf("expected 1 child, but got %d", len(timestampNode.Children))
		}
		if timestampError, ok := timestampNode.Children[0].Value.(string); ok {
			expectedErrorMsg := "error: failed to parse timestamp countersignature: invalid TSTInfo: mismatched message"
			if timestampError != expectedErrorMsg {
				t.Fatalf("expected %s, but got %s", expectedErrorMsg, timestampError)
			}
		} else {
			t.Fatal("expected timestamp node")
		}
	})
}
