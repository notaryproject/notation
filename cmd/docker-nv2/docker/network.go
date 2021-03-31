package docker

import (
	"net/http"

	"github.com/notaryproject/nv2/pkg/registry"
)

// Transport returns the configured round tripper for a host
func Transport(hostname string) (http.RoundTripper, error) {
	tr := http.DefaultTransport
	username, password, err := BasicCredentialFromDockerConfig(hostname)
	if err != nil {
		return nil, err
	}
	if username == "" {
		return tr, nil
	}
	return registry.NewAuthtransport(tr, username, password), nil
}
