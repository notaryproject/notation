package utils

// VirtualHost is a virtualized host machine isolated by environment variable.
type VirtualHost struct {
	Executor     *ExecOpts
	CleanDirFunc func()
	UserDir      string
	env          map[string]string
}

// NewVirtualHost creates a temporary user-level directory and updates
// the "XDG_CONFIG_HOME" environment variable for Executor of the VirtualHost.
func NewVirtualHost(binPath string, options ...Option) (*VirtualHost, error) {
	vhost := &VirtualHost{
		Executor: Binary(binPath),
	}

	var err error
	// setup a temp user directory
	vhost.UserDir, vhost.CleanDirFunc, err = TempUserDir()
	if err != nil {
		return nil, err
	}

	// set user dir environment variables
	vhost.UpdateEnv(UserConfigEnv(vhost.UserDir))

	// set options
	for _, option := range options {
		if err := option(vhost); err != nil {
			panic(err)
		}
	}
	return vhost, nil
}

func (h *VirtualHost) CleanDir() {
	if h.CleanDirFunc != nil {
		h.CleanDirFunc()
	}
}

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

type Option func(vhost *VirtualHost) error

// UserConfigEnv creates environment variable for changing
// user config dir (By setting $XDG_CONFIG_HOME).
func UserConfigEnv(dir string) map[string]string {
	// create and set user dir for linux
	env := make(map[string]string)
	env["XDG_CONFIG_HOME"] = dir
	return env
}
