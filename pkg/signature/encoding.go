package signature

import (
	"encoding/base64"
	"fmt"

	"github.com/docker/go/canonical/json"
)

// EncodeSegment JWT specific base64url encoding with padding stripped
func EncodeSegment(seg []byte) string {
	return base64.RawURLEncoding.EncodeToString(seg)
}

// DecodeSegment JWT specific base64url encoding with padding stripped
func DecodeSegment(seg string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(seg)
}

// DecodeClaims JWT specific base64url encoding with padding stripped as Claims
func DecodeClaims(seg string) (Claims, error) {
	bytes, err := DecodeSegment(seg)
	if err != nil {
		return Claims{}, fmt.Errorf("invalid base64 encoded claims: %v", err)
	}
	var claims Claims
	if err := json.Unmarshal(bytes, &claims); err != nil {
		return Claims{}, fmt.Errorf("invalid JSON encoded claims: %v", err)
	}
	return claims, nil
}
