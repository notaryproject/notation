package signature

import (
	"encoding/json"
)

// Signed is the high level, partially deserialized metadata object
type Signed struct {
	Signed     json.RawMessage `json:"signed"`
	Signatures []Signature     `json:"signatures"`
}

// Content contains the contents to be signed
type Content struct {
	Expiration int64      `json:"exp,omitempty"`
	NotBefore  int64      `json:"nbf,omitempty"`
	IssuedAt   int64      `json:"iat,omitempty"`
	Manifests  []Manifest `json:"manifests"`
}

// Manifest to be signed
type Manifest struct {
	Digest     string   `json:"digest"`
	Size       int64    `json:"size"`
	References []string `json:"references,omitempty"`
}

// Signature to verify the content
type Signature struct {
	Type      string   `json:"typ"`
	Algorithm string   `json:"alg,omitempty"`
	KeyID     string   `json:"kid,omitempty"`
	X5c       [][]byte `json:"x5c,omitempty"`
	Signature []byte   `json:"sig"`
}
