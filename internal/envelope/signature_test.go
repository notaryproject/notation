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
	"testing"

	"github.com/notaryproject/notation-core-go/signature"
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
	val, ok := unsignedAttr["timestampSignature"].(TimestampInfo)
	if !ok {
		t.Fatal("expected to have timestampSignature")
	}
	if val.Error != expectedErrMsg {
		t.Fatalf("expected %s, but got %s", expectedErrMsg, val.Error)
	}
}
