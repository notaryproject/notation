package registry

import (
	"net/http"
)

// Client is a customized registry client
type Client struct {
	base      http.RoundTripper
	plainHTTP bool
}

// NewClient creates a new registry client
func NewClient(base http.RoundTripper, plainHTTP bool) *Client {
	if base == nil {
		base = http.DefaultTransport
	}
	return &Client{
		base:      base,
		plainHTTP: plainHTTP,
	}
}
