package utils

import (
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
)

// VirtualHost is a virtualized host machine isolated by environment variable.
type VirtualHost struct {
	Executor *ExecOpts

	userDir string
	env     map[string]string
}

// NewVirtualHost creates a temporary user-level directory and updates
// the "XDG_CONFIG_HOME" environment variable for Executor of the VirtualHost.
func NewVirtualHost(binPath string, options ...HostOption) (*VirtualHost, error) {
	vhost := &VirtualHost{
		Executor: Binary(binPath),
	}

	var err error
	// setup a temp user directory
	vhost.userDir = ginkgo.GinkgoT().TempDir()
	if err != nil {
		return nil, err
	}

	// set user dir environment variables
	vhost.UpdateEnv(UserConfigEnv(vhost.userDir))

	// set options
	vhost.SetOption(options...)
	return vhost, nil
}

// UserPath returns the path of the absolute path for the given path
// elements that are relative to the user directory.
func (h *VirtualHost) UserPath(elem ...string) string {
	userElem := []string{h.userDir}
	userElem = append(userElem, elem...)
	return filepath.Join(userElem...)
}

// UpdateEnv updates the environment variables for the VirtualHost.
func (h *VirtualHost) UpdateEnv(env map[string]string) {
	if env == nil {
		return
	}
	if h.env == nil {
		h.env = make(map[string]string)
	}
	for key, value := range env {
		h.env[key] = value
	}
	// update ExecOpts.env
	h.Executor.WithEnv(h.env)
}

// SetOption sets the options for the host.
func (h *VirtualHost) SetOption(options ...HostOption) {
	for _, option := range options {
		if err := option(h); err != nil {
			panic(err)
		}
	}
}

// HostOption is a function to set the host configuration.
type HostOption func(vhost *VirtualHost) error

// UserConfigEnv creates environment variable for changing
// user config dir (By setting $XDG_CONFIG_HOME).
func UserConfigEnv(dir string) map[string]string {
	// create and set user dir for linux
	env := make(map[string]string)
	env["XDG_CONFIG_HOME"] = dir
	return env
}
