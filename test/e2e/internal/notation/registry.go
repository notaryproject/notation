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

func (r *Artifact) GUN() string {
	return fmt.Sprintf("%s/%s", r.Host, r.Reference())
}

func (r *Artifact) Reference() string {
	return fmt.Sprintf("%s:%s", r.Repo, r.Tag)
}

func (r *Artifact) ClearImage() error {
	return os.RemoveAll(filepath.Join(OCILayoutPath, r.Repo))
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

// GenImage generates a new image with a new repository.
func GenImage() *Artifact {
	// generate new newRepo
	newRepo := fmt.Sprintf("%s%d", testRepo, genRepoId())

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

func genRepoId() int {
	var id int

	repoIdMu.Lock()
	id = repoId
	repoId++
	repoIdMu.Unlock()

	return id
}
