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
