package signature

import "github.com/notaryproject/nv2/pkg/reference"

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

// DescriptorFromReference converts descriptor from generic reference
func DescriptorFromReference(d reference.Descriptor) Descriptor {
	result := Descriptor{
		MediaType: d.MediaType,
		Size:      d.Size,
	}
	if len(d.Digests) > 0 {
		result.Digest = d.Digests[0].String()
	}
	return result
}
