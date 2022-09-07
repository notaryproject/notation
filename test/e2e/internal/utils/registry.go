package utils

import (
	"crypto/rand"
	"fmt"
	"net/url"
	"path"

	oregistry "oras.land/oras-go/v2/registry"
)

type registry struct {
	Host     string
	Username string
	Password string
	Artifact string
}

// Validate validates the registry and artifact is valid.
func (r *registry) Validate() error {
	if _, err := url.ParseRequestURI(r.Host); err != nil {
		return err
	}
	ref, err := oregistry.ParseReference(r.Artifact)
	if err != nil {
		return err
	}
	if ref.Registry != r.Host {
		return fmt.Errorf("registry host %q mismatch base image %q", r.Host, r.Artifact)
	}
	return nil
}

// NewRegistry creates a new registry with given parameters.
func NewRegistry(host, username, password, artifact string) *registry {
	return &registry{
		Host:     host,
		Username: username,
		Password: password,
		Artifact: artifact,
	}
}

func (r *registry) tagImage(repo, version string) (string, error) {
	target := path.Join(r.Host, repo) + ":" + version
	args := []string{"tag", r.Artifact, target}
	_, err := Exec("docker", ExecOpts{}, args...)
	if err != nil {
		return "", err
	}
	return target, nil
}

func (r *registry) rmImage(ref string) error {
	_, err := Exec("docker", ExecOpts{}, "rmi", ref)
	if err != nil {
		return err
	}
	return nil
}

// PushImage push an image [host/repo:version] to the registry.
func (r *registry) PushImage(repo, version string) (newImage string, cleaner func(), err error) {
	r.Login()
	defer r.Logout()

	newImage, err = r.tagImage(repo, version)
	if err != nil {
		return
	}
	args := []string{"push", newImage}
	_, err = Exec("docker", ExecOpts{}, args...)
	if err != nil {
		return
	}
	cleaner = func() {
		r.rmImage(newImage)
	}
	return newImage, cleaner, nil
}

// PushRandomImage tags an image and push it to the registry. The image name is generated randomly.
func (r *registry) PushRandomImage() (newImage string, cleaner func(), err error) {
	b := make([]byte, 16)
	rand.Read(b)
	repo := fmt.Sprintf("%x", b)
	return r.PushImage(repo, "v1")
}

// Login logins to the registry by calling docker login.
func (r *registry) Login() error {
	args := []string{"login", r.Host, "-u", r.Username, "-p", r.Password}
	_, err := Exec("docker", ExecOpts{}, args...)
	return err
}

// Logout logouts the registry by calling docker logout.
func (r *registry) Logout() error {
	args := []string{"logout", r.Host}
	_, err := Exec("docker", ExecOpts{}, args...)
	return err
}
