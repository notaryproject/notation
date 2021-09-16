package config

import "sync"

var (
	instance     *File
	instanceOnce sync.Once
)

// LoadOrDefaultOnce returns the previously read config file.
// If previous config file does not exists, it reads the config from file
// or return a default config if not found.
// The returned config is only suitable for read only scenarios for short-lived processes.
func LoadOrDefaultOnce() (*File, error) {
	var err error
	instanceOnce.Do(func() {
		instance, err = LoadOrDefault()
	})
	return instance, err
}
