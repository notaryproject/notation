package notation

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
)

// OCILayout is a OCI layout directory for
type OCILayout struct {
	// Path is the path of the OCI layout directory.
	Path string
	// Tag is the tag of artifact in the OCI layout.
	Tag string
	// Digest is the digest of artifact in the OCI layout.
	Digest string
}

// GenerateOCILayout creates a new OCI layout in a temporary directory.
func GenerateOCILayout(srcRepoName string) (*OCILayout, error) {
	ctx := context.Background()

	if srcRepoName == "" {
		srcRepoName = TestRepoUri
	}

	destPath := filepath.Join(GinkgoT().TempDir(), newRepoName())
	// create a local store from OCI layout directory.
	srcStore, err := oci.NewFromFS(ctx, os.DirFS(filepath.Join(OCILayoutPath, srcRepoName)))
	if err != nil {
		return nil, err
	}

	// create a dest store for store the generated oci layout.
	destStore, err := oci.New(destPath)
	if err != nil {
		return nil, err
	}

	// copy data
	desc, err := oras.ExtendedCopy(ctx, srcStore, TestTag, destStore, "", oras.DefaultExtendedCopyOptions)
	if err != nil {
		return nil, err
	}
	return &OCILayout{
		Path:   destPath,
		Tag:    TestTag,
		Digest: desc.Digest.String(),
	}, nil
}

// ReferenceWithTag returns the reference with tag.
func (o *OCILayout) ReferenceWithTag() string {
	return o.Path + ":" + o.Tag
}

// ReferenceWithDigest returns the reference with digest.
func (o *OCILayout) ReferenceWithDigest() string {
	return o.Path + "@" + o.Digest
}
