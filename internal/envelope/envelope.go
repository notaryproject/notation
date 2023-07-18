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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-core-go/signature/cose"
	"github.com/notaryproject/notation-core-go/signature/jws"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	// Supported envelope format.
	COSE = "cose"
	JWS  = "jws"

	// MediaTypePayloadV1 is the supported content type for signature's payload.
	MediaTypePayloadV1 = "application/vnd.cncf.notary.payload.v1+json"
)

// Payload describes the content that gets signed.
type Payload struct {
	TargetArtifact ocispec.Descriptor `json:"targetArtifact"`
}

// GetEnvelopeMediaType converts the envelope type to mediaType name.
func GetEnvelopeMediaType(sigFormat string) (string, error) {
	switch sigFormat {
	case JWS:
		return jws.MediaTypeEnvelope, nil
	case COSE:
		return cose.MediaTypeEnvelope, nil
	}
	return "", fmt.Errorf("signature format %q not supported", sigFormat)
}

// ValidatePayloadContentType validates signature payload's content type.
func ValidatePayloadContentType(payload *signature.Payload) error {
	switch payload.ContentType {
	case MediaTypePayloadV1:
		return nil
	default:
		return fmt.Errorf("payload content type %q not supported", payload.ContentType)
	}
}

// DescriptorFromPayload parses a signature payload and returns the descriptor
// that was signed. Note: the descriptor was signed but may not be trusted
func DescriptorFromSignaturePayload(payload *signature.Payload) (*ocispec.Descriptor, error) {
	if payload == nil {
		return nil, errors.New("empty payload")
	}

	err := ValidatePayloadContentType(payload)
	if err != nil {
		return nil, err
	}

	var parsedPayload Payload
	err = json.Unmarshal(payload.Content, &parsedPayload)
	if err != nil {
		return nil, errors.New("failed to unmarshall the payload content to Payload")
	}

	return &parsedPayload.TargetArtifact, nil
}
