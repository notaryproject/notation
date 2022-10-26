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
		Usage:     "username for registry operations (if not specified, defaults to $NOTATION_USERNAME)",
	}
	setflagUsername = func(fs *pflag.FlagSet, p *string) {
		fs.StringVarP(p, flagUsername.Name, flagUsername.Shorthand, "", flagUsername.Usage)
	}

	flagPassword = &pflag.Flag{
		Name:      "password",
		Shorthand: "p",
		Usage:     "password for registry operations (if not specified, defaults to $NOTATION_PASSWORD)",
	}
	setFlagPassword = func(fs *pflag.FlagSet, p *string) {
		fs.StringVarP(p, flagPassword.Name, flagPassword.Shorthand, "", flagPassword.Usage)
	}

	flagPlainHTTP = &pflag.Flag{
		Name:     "plain-http",
		Usage:    "Registry access via plain HTTP",
		DefValue: "false",
	}
	setFlagPlainHTTP = func(fs *pflag.FlagSet, p *bool) {
		fs.BoolVar(p, flagPlainHTTP.Name, false, flagPlainHTTP.Usage)
	}

	flagMediaType = &pflag.Flag{
		Name:     "media-type",
		Usage:    "specify the media type of the manifest read from file or stdin",
		DefValue: defaultMediaType,
	}
	setFlagMediaType = func(fs *pflag.FlagSet, p *string) {
		fs.StringVar(p, flagMediaType.Name, defaultMediaType, flagMediaType.Usage)
	}

	flagLocal = &pflag.Flag{
		Name:      "local",
		Shorthand: "l",
		Usage:     "reference is a local file",
		DefValue:  "false",
	}
	setFlagLocal = func(fs *pflag.FlagSet, p *bool) {
		fs.BoolVarP(p, flagLocal.Name, flagLocal.Shorthand, false, flagLocal.Usage)
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

type CommonFlagOpts struct {
	Local     bool
	MediaType string
}

// ApplyFlags set flags and their default values for the FlagSet
func (opts *CommonFlagOpts) ApplyFlags(fs *pflag.FlagSet) {
	setFlagMediaType(fs, &opts.MediaType)
	setFlagLocal(fs, &opts.Local)
}

type RemoteFlagOpts struct {
	SecureFlagOpts
	CommonFlagOpts
}

// ApplyFlags set flags and their default values for the FlagSet
func (opts *RemoteFlagOpts) ApplyFlags(fs *pflag.FlagSet) {
	opts.SecureFlagOpts.ApplyFlags(fs)
	opts.CommonFlagOpts.ApplyFlags(fs)
}
