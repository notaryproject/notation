package envelope

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/notaryproject/notation-core-go/signature/cose"
	"github.com/notaryproject/notation-core-go/signature/jws"
	gcose "github.com/veraison/go-cose"
)

var (
	validJwsSignatureEnvelope, _ = json.Marshal(struct{}{})
	validCoseSignatureEnvelope   []byte
	invalidSignatureEnvelope     = []byte("invalid")
)

func init() {
	msg := gcose.Sign1Message{
		Headers:   gcose.NewSign1Message().Headers,
		Payload:   []byte("valid"),
		Signature: []byte("valid"),
	}
	validCoseSignatureEnvelope, _ = msg.MarshalCBOR()
}

func checkErrorEqual(expected, got error) bool {
	if expected == nil && got == nil {
		return true
	}
	if expected != nil && got != nil {
		return expected.Error() == got.Error()
	}
	return false
}

func TestSpeculateSignatureEnvelopeFormat(t *testing.T) {
	tests := []struct {
		name         string
		raw          []byte
		expectedType string
		expectedErr  error
	}{
		{
			name:         "jws signature media type",
			raw:          validJwsSignatureEnvelope,
			expectedType: jws.MediaTypeEnvelope,
			expectedErr:  nil,
		},
		{
			name:         "cose signature media type",
			raw:          validCoseSignatureEnvelope,
			expectedType: cose.MediaTypeEnvelope,
			expectedErr:  nil,
		},
		{
			name:         "invalid signature media type",
			raw:          invalidSignatureEnvelope,
			expectedType: "",
			expectedErr:  errors.New("unsupported signature format"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eType, err := SpeculateSignatureEnvelopeFormat(tt.raw)
			if !checkErrorEqual(tt.expectedErr, err) {
				t.Fatalf("expected speculate signature envelope format err: %v, got: %v", tt.expectedErr, err)
			}
			if eType != tt.expectedType {
				t.Fatalf("expected signature envelopeType: %v, got: %v", tt.expectedType, eType)
			}
		})
	}
}

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
