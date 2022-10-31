package main

import (
	"os"

	"github.com/spf13/pflag"
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

	flagPlainHTTP = &pflag.Flag{
		Name:     "plain-http",
		Usage:    "registry access via plain HTTP",
		DefValue: "false",
	}
	setFlagPlainHTTP = func(fs *pflag.FlagSet, p *bool) {
		fs.BoolVar(p, flagPlainHTTP.Name, false, flagPlainHTTP.Usage)
	}
)

type SecureFlagOpts struct {
	Username  string
	Password  string
	PlainHTTP bool
}

// ApplyFlags set flags and their default values for the FlagSet
func (opts *SecureFlagOpts) ApplyFlags(fs *pflag.FlagSet) {
	setflagUsername(fs, &opts.Username)
	setFlagPassword(fs, &opts.Password)
	setFlagPlainHTTP(fs, &opts.PlainHTTP)
	opts.Username = os.Getenv(defaultUsernameEnv)
	opts.Password = os.Getenv(defaultPasswordEnv)
}
