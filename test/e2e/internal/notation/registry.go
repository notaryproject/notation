package notation

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync/atomic"

	oregistry "oras.land/oras-go/v2/registry"
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

var TestRegistry = Registry{}

type Artifact struct {
	*Registry
	Repo string
	Tag  string
}

// GenerateArtifact generates a new image with a new repository.
func GenerateArtifact() *Artifact {
	// generate new newRepo
	newRepo := fmt.Sprintf("%s-%d", testRepo, genRepoId())

	// copy oci layout to the new repo
	if err := copyDir(filepath.Join(OCILayoutPath, testRepo), filepath.Join(RegistryStoragePath, newRepo)); err != nil {
		panic(err)
	}

	image := &Artifact{
		Registry: &Registry{
			Host:     TestRegistry.Host,
			Username: TestRegistry.Username,
			Password: TestRegistry.Password,
		},
		Repo: newRepo,
		Tag:  "v1",
	}
	if err := image.Validate(); err != nil {
		panic(err)
	}
	return image
}

// Validate validates the registry and artifact is valid.
func (r *Artifact) Validate() error {
	if _, err := url.ParseRequestURI(r.Host); err != nil {
		return err
	}
	ref, err := oregistry.ParseReference(r.Reference())
	if err != nil {
		return err
	}
	if ref.Registry != r.Host {
		return fmt.Errorf("registry host %q mismatch base image %q", r.Host, r.Repo)
	}
	return nil
}

// Reference returns the <registryHost>/<Repository>:<Tag>
func (r *Artifact) Reference() string {
	return fmt.Sprintf("%s/%s:%s", r.Host, r.Repo, r.Tag)
}

// Reference removes the the repository of the artifact.
func (r *Artifact) Remove() error {
	return os.RemoveAll(filepath.Join(OCILayoutPath, r.Repo))
}

// genRepoId returns a new repoId
func genRepoId() int64 {
	return atomic.AddInt64(&repoId, 1)
}
