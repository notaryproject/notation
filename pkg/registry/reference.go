package registry

import (
	"net/url"
	"strings"
)

// Reference references to a descriptor in the registry
type Reference struct {
	// Registry is the name of the registry.
	// It is usually the domain name of the registry.
	Registry string

	// Repository is the name of the repository
	Repository string

	// Reference is the reference of the object in the repository.
	// A reference can be a tag and / or a digest.
	Reference string
}

func ParseReferenceFromURL(uri *url.URL) Reference {
	var repository string
	var reference string
	path := strings.TrimPrefix(uri.Path, "/")
	if index := strings.Index(path, "@"); index != -1 {
		repository = path[:index]
		reference = path[index+1:]
	} else if index := strings.Index(path, ":"); index != -1 {
		repository = path[:index]
		reference = path[index+1:]
	} else {
		repository = path
	}
	return Reference{
		Registry:   uri.Host,
		Repository: repository,
		Reference:  reference,
	}
}

// Host returns the host name of the registry
func (r Reference) Host() string {
	if r.Registry == "docker.io" {
		return "registry-1.docker.io"
	}
	return r.Registry
}

// ReferenceOrDefault returns the reference or the default reference if empty.
func (r Reference) ReferenceOrDefault() string {
	if r.Reference == "" {
		return "latest"
	}
	return r.Reference
}
