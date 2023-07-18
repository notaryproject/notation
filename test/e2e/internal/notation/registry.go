// Copyright The Notary Project Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package notation

import (
	"context"
	"fmt"
	"hash/maphash"
	"net"
	"os"
	"path/filepath"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

const ArtifactTypeNotation = "application/vnd.cncf.notary.signature"

type Registry struct {
	// Host is the registry host.
	Host string
	// Username is the username to access the registry.
	Username string
	// Password is the password to access the registry.
	Password string
	// DomainHost is a registry host, separate from localhost, used for testing
	// the --insecure-registry flag.
	//
	// If the host is localhost, Notation connects via plain HTTP. For
	// non-localhost hosts, Notation defaults to HTTPS. However, users can
	// enforce HTTP by setting the --insecure-registry flag.
	DomainHost string
}

// CreateArtifact copies a local OCI layout to the registry to create
// a new artifact with a new repository.
//
// srcRepoName is the repo name in ./testdata/registry/oci_layout folder.
// destRepoName is the repo name to be created in the registry.
func (r *Registry) CreateArtifact(srcRepoName, destRepoName string) (*Artifact, error) {
	ctx := context.Background()
	// create a local store from OCI layout directory.
	srcStore, err := oci.NewFromFS(ctx, os.DirFS(filepath.Join(OCILayoutPath, srcRepoName)))
	if err != nil {
		return nil, err
	}

	// create the artifact struct
	artifact := &Artifact{
		Registry: r,
		Repo:     destRepoName,
		Tag:      TestTag,
	}

	// create the remote.repository
	destRepo, err := newRepository(artifact.ReferenceWithTag())
	if err != nil {
		return nil, err
	}

	// copy data
	desc, err := oras.ExtendedCopy(ctx, srcStore, artifact.Tag, destRepo, "", oras.DefaultExtendedCopyOptions)
	if err != nil {
		return nil, err
	}
	artifact.Digest = desc.Digest.String()
	return artifact, err
}

var TestRegistry = Registry{}

// Artifact describes an artifact in a repository.
type Artifact struct {
	*Registry
	// Repo is the repository name.
	Repo string
	// Tag is the tag of the artifact.
	Tag string
	// Digest is the digest of the artifact.
	Digest string
}

// GenerateArtifact generates a new artifact with a new repository by copying
// the source repository in the OCILayoutPath to be a new repository.
func GenerateArtifact(srcRepo, newRepo string) *Artifact {
	if srcRepo == "" {
		srcRepo = TestRepoUri
	}

	if newRepo == "" {
		// generate new repo
		newRepo = newRepoName()
	}

	artifact, err := TestRegistry.CreateArtifact(srcRepo, newRepo)
	if err != nil {
		panic(err)
	}
	return artifact
}

// ReferenceWithTag returns the <registryHost>/<Repository>:<Tag>
func (r *Artifact) ReferenceWithTag() string {
	return fmt.Sprintf("%s/%s:%s", r.Host, r.Repo, r.Tag)
}

// ReferenceWithDigest returns the <registryHost>/<Repository>@<alg>:<digest>
func (r *Artifact) ReferenceWithDigest() string {
	return fmt.Sprintf("%s/%s@%s", r.Host, r.Repo, r.Digest)
}

// DomainReferenceWithDigest returns the <domainHost>/<Repository>@<alg>:<digest>
// for testing --insecure-registry flag and TLS request.
func (r *Artifact) DomainReferenceWithDigest() string {
	return fmt.Sprintf("%s/%s@%s", r.DomainHost, r.Repo, r.Digest)
}

// SignatureManifest returns the manifest of the artifact.
func (r *Artifact) SignatureDescriptors() ([]ocispec.Descriptor, error) {
	ctx := context.Background()
	repo, err := newRepository(r.ReferenceWithDigest())
	if err != nil {
		return nil, err
	}

	// get manifest descriptor
	desc, err := repo.Manifests().Resolve(ctx, r.ReferenceWithDigest())
	if err != nil {
		return nil, err
	}

	// get signature descriptors
	var descriptors []ocispec.Descriptor
	if err := repo.Referrers(context.Background(), desc, ArtifactTypeNotation, func(referrers []ocispec.Descriptor) error {
		descriptors = append(descriptors, referrers...)
		return nil
	}); err != nil {
		return nil, err
	}
	return descriptors, nil
}

func newRepoName() string {
	var newRepo string
	seed := maphash.MakeSeed()
	newRepo = fmt.Sprintf("%s-%d", TestRepoUri, maphash.Bytes(seed, nil))
	return newRepo
}

func newRepository(reference string) (*remote.Repository, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, err
	}

	repo := &remote.Repository{
		Client:    authClient(ref),
		Reference: ref,
		PlainHTTP: false,
	}
	if host, _, _ := net.SplitHostPort(ref.Host()); host == "localhost" {
		repo.PlainHTTP = true
	}

	return repo, nil
}

func authClient(ref registry.Reference) *auth.Client {
	return &auth.Client{
		Credential: func(ctx context.Context, registry string) (auth.Credential, error) {
			switch registry {
			case ref.Host():
				return auth.Credential{
					Username: TestRegistry.Username,
					Password: TestRegistry.Password,
				}, nil
			default:
				return auth.EmptyCredential, nil
			}
		},
		Cache:    auth.NewCache(),
		ClientID: "notation",
	}
}
