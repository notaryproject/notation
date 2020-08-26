package signature

// Header defines the signature header
type Header struct {
	Raw  []byte `json:"-"`
	Type string `json:"typ"`
}

// Claims contains the claims to be signed
type Claims struct {
	Manifest
	Expiration int64 `json:"exp,omitempty"`
	IssuedAt   int64 `json:"iat,omitempty"`
	NotBefore  int64 `json:"nbf,omitempty"`
}

// Manifest to be signed
type Manifest struct {
	Descriptor
	References []string `json:"references,omitempty"`
}

// Descriptor describes the basic information of the target content
type Descriptor struct {
	MediaType string `json:"mediaType,omitempty"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}
