package docker

import (
	"net/http"

	"github.com/notaryproject/notary/v2/util"
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
	return util.TransportWithBasicAuth(tr, hostname, username, password), nil
}
