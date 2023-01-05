package notation

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"time"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

type Registry struct {
	Host     string
	Username string
	Password string
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

func newRepoName() string {
	var newRepo string
	for {
		// set the seed with nanosecond precision.
		rand.Seed(time.Now().UnixNano())
		newRepo = fmt.Sprintf("%s-%d", TestRepoUri, rand.Intn(math.MaxInt32))

		_, err := os.Stat(filepath.Join(RegistryStoragePath, newRepo))
		if err != nil {
			if os.IsNotExist(err) {
				// newRepo doesn't exist.
				break
			}
			panic(err)
		}
	}
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
