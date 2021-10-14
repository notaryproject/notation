package registry

import (
	"github.com/notaryproject/notation-go-lib"
	"github.com/opencontainers/go-digest"
)

// DescriptorFromBytes computes the basic descriptor from the given bytes
func DescriptorFromBytes(data []byte) notation.Descriptor {
	return notation.Descriptor{
		Digest: digest.FromBytes(data),
		Size:   int64(len(data)),
	}
}
