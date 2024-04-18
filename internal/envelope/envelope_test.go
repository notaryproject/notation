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

func TestGetEnvelopeMediaType(t *testing.T) {
	type args struct {
		sigFormat string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "jws",
			args:    args{"jws"},
			want:    "application/jose+json",
			wantErr: false,
		},
		{
			name:    "cose",
			args:    args{"cose"},
			want:    "application/cose",
			wantErr: false,
		},
		{
			name:    "unsupported",
			args:    args{"unsupported"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEnvelopeMediaType(tt.args.sigFormat)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEnvelopeMediaType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetEnvelopeMediaType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidatePayloadContentType(t *testing.T) {
	tests := []struct {
		name    string
		payload *signature.Payload
		wantErr bool
	}{
		{
			name: "valid content type",
			payload: &signature.Payload{
				ContentType: MediaTypePayloadV1,
			},
			wantErr: false,
		},
		{
			name: "invalid content type",
			payload: &signature.Payload{
				ContentType: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePayloadContentType(tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePayloadContentType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDescriptorFromSignaturePayload(t *testing.T) {
	validPayload := &signature.Payload{
		ContentType: MediaTypePayloadV1,
		Content:     []byte(`{"targetArtifact": {"mediaType": "application/vnd.oci.image.manifest.v1+json", "size": 314159, "digest": "sha256:abcd1234", "urls": ["http://example.com"]}}`),
	}
	invalidPayload := &signature.Payload{
		ContentType: "invalid",
		Content:     []byte(`invalid`),
	}

	tests := []struct {
		name    string
		payload *signature.Payload
		wantErr bool
	}{
		{
			name:    "valid payload",
			payload: validPayload,
			wantErr: false,
		},
		{
			name:    "invalid content type",
			payload: invalidPayload,
			wantErr: true,
		},
		{
			name:    "nil payload",
			payload: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DescriptorFromSignaturePayload(tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("DescriptorFromSignaturePayload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
