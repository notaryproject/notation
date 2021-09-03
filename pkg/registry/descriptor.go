package registry

import (
	"github.com/notaryproject/notary/v2/signature"
	digest "github.com/opencontainers/go-digest"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

func OCIDescriptorFromNotary(desc signature.Descriptor) oci.Descriptor {
	return oci.Descriptor{
		MediaType: desc.MediaType,
		Digest:    digest.Digest(desc.Digest),
		Size:      desc.Size,
	}
}
