package notation

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync/atomic"

	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

var (
	repoId int64 = 0
)

const (
	testRepo = "e2e"
	testTag  = "v1"
)

type Registry struct {
	Host     string
	Username string
	Password string
}

var TestRegistry Registry

type Artifact struct {
	*Registry
	Repo   string
	Tag    string
	Digest string
}

// GenerateArtifact generates a new image with a new repository.
func GenerateArtifact() *Artifact {
	// generate new newRepo
	newRepo := fmt.Sprintf("%s-%d", testRepo, genRepoId())

	// copy oci layout to the new repo
	if err := copyDir(filepath.Join(OCILayoutPath, testRepo), filepath.Join(RegistryStoragePath, newRepo)); err != nil {
		panic(err)
	}

	artifact := &Artifact{
		Registry: &Registry{
			Host:     TestRegistry.Host,
			Username: TestRegistry.Username,
			Password: TestRegistry.Password,
		},
		Repo: newRepo,
		Tag:  "v1",
	}

	if err := artifact.Validate(); err != nil {
		panic(err)
	}

	if err := artifact.fetchDigest(); err != nil {
		panic(err)
	}

	return artifact
}

// Validate validates the registry and artifact is valid.
func (r *Artifact) Validate() error {
	if _, err := url.ParseRequestURI(r.Host); err != nil {
		return err
	}
	ref, err := registry.ParseReference(r.ReferenceWithTag())
	if err != nil {
		return err
	}
	if ref.Registry != r.Host {
		return fmt.Errorf("registry host %q mismatch base image %q", r.Host, r.Repo)
	}
	return nil
}

func (r *Artifact) fetchDigest() error {
	// create repository
	ref, err := registry.ParseReference(r.ReferenceWithTag())
	if err != nil {
		return err

	}
	authClient := &auth.Client{
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
	repo := &remote.Repository{
		Client:    authClient,
		Reference: ref,
		PlainHTTP: true,
	}

	// resolve descriptor
	descriptor, err := repo.Resolve(context.Background(), r.ReferenceWithTag())
	if err != nil {
		return err
	}

	// set digest
	r.Digest = descriptor.Digest.String()
	return nil
}

// ReferenceWithTag returns the <registryHost>/<Repository>:<Tag>
func (r *Artifact) ReferenceWithTag() string {
	return fmt.Sprintf("%s/%s:%s", r.Host, r.Repo, r.Tag)
}

// ReferenceWithDigest returns the <registryHost>/<Repository>@<alg>:<digest>
func (r *Artifact) ReferenceWithDigest() string {
	return fmt.Sprintf("%s/%s@%s", r.Host, r.Repo, r.Digest)
}

// Reference removes the the repository of the artifact.
func (r *Artifact) Remove() error {
	return os.RemoveAll(filepath.Join(OCILayoutPath, r.Repo))
}

// genRepoId returns a new repoId
func genRepoId() int64 {
	return atomic.AddInt64(&repoId, 1)
}
