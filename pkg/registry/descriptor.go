package registry

import (
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// DescriptorFromBytes computes the basic descriptor from the given bytes
func DescriptorFromBytes(data []byte) ocispec.Descriptor {
	return ocispec.Descriptor{
		Digest: digest.FromBytes(data),
		Size:   int64(len(data)),
	}
}
