package main

import (
	"os"

	"github.com/spf13/pflag"
	"oras.land/oras-go/v2/registry/remote/auth"
)

const (
	defaultUsernameEnv = "NOTATION_USERNAME"
	defaultPasswordEnv = "NOTATION_PASSWORD"
	defaultMediaType   = "application/vnd.docker.distribution.manifest.v2+json"
)

var (
	flagUsername = &pflag.Flag{
		Name:      "username",
		Shorthand: "u",
		Usage:     "username for registry operations (default to $NOTATION_USERNAME if not specified)",
	}
	setflagUsername = func(fs *pflag.FlagSet, p *string) {
		fs.StringVarP(p, flagUsername.Name, flagUsername.Shorthand, "", flagUsername.Usage)
	}

	flagPassword = &pflag.Flag{
		Name:      "password",
		Shorthand: "p",
		Usage:     "password for registry operations (default to $NOTATION_PASSWORD if not specified)",
	}
	setFlagPassword = func(fs *pflag.FlagSet, p *string) {
		fs.StringVarP(p, flagPassword.Name, flagPassword.Shorthand, "", flagPassword.Usage)
	}

	flagInsecureRegistry = &pflag.Flag{
		Name:     "insecure-registry",
		Usage:    "use HTTP protocol while connecting to registries. Should be used only for testing",
		DefValue: "false",
	}
	setFlagInsecureRegistry = func(fs *pflag.FlagSet, p *bool) {
		fs.BoolVar(p, flagInsecureRegistry.Name, false, flagInsecureRegistry.Usage)
	}
)

type SecureFlagOpts struct {
	Username         string
	Password         string
	InsecureRegistry bool
}

// ApplyFlags set flags and their default values for the FlagSet
func (opts *SecureFlagOpts) ApplyFlags(fs *pflag.FlagSet) {
	setflagUsername(fs, &opts.Username)
	setFlagPassword(fs, &opts.Password)
	setFlagInsecureRegistry(fs, &opts.InsecureRegistry)
	opts.Username = os.Getenv(defaultUsernameEnv)
	opts.Password = os.Getenv(defaultPasswordEnv)
}

// Credential returns an auth.Credential from opts.Username and opts.Password.
func (opts *SecureFlagOpts) Credential() auth.Credential {
	if opts.Username == "" {
		return auth.Credential{
			RefreshToken: opts.Password,
		}
	}
	return auth.Credential{
		Username: opts.Username,
		Password: opts.Password,
	}
}
