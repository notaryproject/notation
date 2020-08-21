package reference

import "github.com/opencontainers/go-digest"

// Descriptor describes the basic information of the target content
type Descriptor struct {
	MediaType string          `json:"mediaType,omitempty"`
	Digests   []digest.Digest `json:"digests"`
	Size      int64           `json:"size"`
}
