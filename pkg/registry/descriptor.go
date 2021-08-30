package registry

import (
	"github.com/notaryproject/notation-go-lib/signature"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func OCIDescriptorFromNotation(desc signature.Descriptor) ocispec.Descriptor {
	return ocispec.Descriptor{
		MediaType: desc.MediaType,
		Digest:    digest.Digest(desc.Digest),
		Size:      desc.Size,
	}
}
