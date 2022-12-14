package envelope

import (
	"testing"

	gcose "github.com/veraison/go-cose"
)

var validCoseSignatureEnvelope []byte

func init() {
	msg := gcose.Sign1Message{
		Headers:   gcose.NewSign1Message().Headers,
		Payload:   []byte("valid"),
		Signature: []byte("valid"),
	}
	validCoseSignatureEnvelope, _ = msg.MarshalCBOR()
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
