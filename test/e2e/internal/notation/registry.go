package notation

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	oregistry "oras.land/oras-go/v2/registry"
)

var (
	repoId   = 0
	repoIdMu = sync.Mutex{}
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

// GenArtifact generates a new image with a new repository.
func GenArtifact() *Artifact {
	// generate new newRepo
	newRepo := fmt.Sprintf("%s-%d", testRepo, genRepoId())

	// copy oci layout to the new repo
	copyDir(
		filepath.Join(OCILayoutPath, testRepo),
		filepath.Join(RegistryStoragePath, newRepo))

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
	ref, err := oregistry.ParseReference(r.GUN())
	if err != nil {
		return err
	}
	if ref.Registry != r.Host {
		return fmt.Errorf("registry host %q mismatch base image %q", r.Host, r.Repo)
	}
	return nil
}

// GUN returns the <registryHost>/<Repository>/<Tag>
func (r *Artifact) GUN() string {
	return fmt.Sprintf("%s/%s", r.Host, r.Reference())
}

// Reference return the <Repository>/<tag>
func (r *Artifact) Reference() string {
	return fmt.Sprintf("%s:%s", r.Repo, r.Tag)
}

// Reference removes the the repository of the artifact.
func (r *Artifact) Remove() error {
	return os.RemoveAll(filepath.Join(OCILayoutPath, r.Repo))
}

// genRepoId returns a new repoId
func genRepoId() int {
	var id int

	repoIdMu.Lock()
	id = repoId
	repoId++
	repoIdMu.Unlock()

	return id
}
