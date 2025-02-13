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

package json

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	coresignature "github.com/notaryproject/notation-core-go/signature"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type errorEnvelope struct{}

func (e errorEnvelope) Sign(req *coresignature.SignRequest) ([]byte, error) {
	return nil, errors.New("mock sign error")
}

func (e errorEnvelope) Verify() (*coresignature.EnvelopeContent, error) {
	return nil, errors.New("mock verify error")
}

func (e errorEnvelope) Content() (*coresignature.EnvelopeContent, error) {
	return nil, errors.New("mock content error")
}

func TestGetUnsignedAttributes(t *testing.T) {
	envContent := &coresignature.EnvelopeContent{
		SignerInfo: coresignature.SignerInfo{
			UnsignedAttributes: coresignature.UnsignedAttributes{
				TimestampSignature: []byte("invalid"),
			},
		},
	}
	expectedErrMsg := "failed to parse timestamp countersignature: cms: syntax error: invalid signed data: failed to convert from BER to DER: asn1: syntax error: decoding BER length octets: short form length octets value should be less or equal to the subsequent octets length"
	unsignedAttr := getUnsignedAttributes(envContent)
	val, ok := unsignedAttr["timestampSignature"].(*timestamp)
	if !ok {
		t.Fatal("expected to have timestampSignature")
	}
	if val.Error != expectedErrMsg {
		t.Fatalf("expected %s, but got %s", expectedErrMsg, val.Error)
	}
}

func TestGetSignedAttributes(t *testing.T) {
	expiry := time.Now()
	envContent := &coresignature.EnvelopeContent{
		SignerInfo: coresignature.SignerInfo{
			SignedAttributes: coresignature.SignedAttributes{
				Expiry: expiry,
				ExtendedAttributes: []coresignature.Attribute{
					{
						Key:   "keyName",
						Value: "value",
					},
				},
			},
		},
	}
	signedAttr := getSignedAttributes(envContent)
	if signedAttr["expiry"] != expiry {
		t.Fatalf("expected %s, but got %s", expiry, signedAttr["expiry"])
	}

	if signedAttr["keyName"] != "value" {
		t.Fatalf("expected value, but got %s", signedAttr["keyName"])
	}
}

func TestParseTimestamp(t *testing.T) {
	t.Run("invalid timestamp signature", func(t *testing.T) {
		signerInfo := coresignature.SignerInfo{
			UnsignedAttributes: coresignature.UnsignedAttributes{
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
		tsaToken, err := os.ReadFile("../testdata/TimeStampTokenWithInvalidSignature.p7s")
		if err != nil {
			t.Fatal(err)
		}

		signerInfo := coresignature.SignerInfo{
			UnsignedAttributes: coresignature.UnsignedAttributes{
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

func TestInspectSignature_NewSignatureError(t *testing.T) {
	h := NewInspectHandler(nil)
	// ...existing code to ensure h.output.MediaType is set...
	h.OnReferenceResolved("test-ref", "test-media-type")
	manifestDesc := ocispec.Descriptor{Digest: "fake-digest"}
	sigDesc := ocispec.Descriptor{MediaType: "fake-media-type"}

	err := h.InspectSignature(manifestDesc, sigDesc, errorEnvelope{})
	if err == nil || !strings.Contains(err.Error(), "mock content error") {
		t.Fatalf("expected error 'mock content error', got %v", err)
	}
}
