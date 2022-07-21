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
		Usage:     "Username for registry operations (default from $NOTATION_USERNAME)",
	}
	setflagUsername = func(fs *pflag.FlagSet, p *string) {
		fs.StringVarP(p, flagUsername.Name, flagUsername.Shorthand, "", flagUsername.Usage)
	}

	flagPassword = &pflag.Flag{
		Name:      "password",
		Shorthand: "p",
		Usage:     "Password for registry operations (default from $NOTATION_PASSWORD)",
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

	flagOutput = &pflag.Flag{
		Name:      "output",
		Shorthand: "o",
		Usage:     "write signature to a specific path",
	}
	setFlagOutput = func(fs *pflag.FlagSet, p *string) {
		fs.StringVarP(p, flagOutput.Name, flagOutput.Shorthand, "", flagOutput.Usage)
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

	flagSignature = &pflag.Flag{
		Name:      "signature",
		Shorthand: "s",
		Usage:     "signature files",
	}
	setFlagSignature = func(fs *pflag.FlagSet, p *[]string) {
		fs.StringSliceVarP(p, flagSignature.Name, flagSignature.Shorthand, []string{}, flagSignature.Usage)
	}
)

type SecureFlagOpts struct {
	Username  string
	Password  string
	PlainHTTP bool
}

func (opts *SecureFlagOpts) ApplyFlag(fs *pflag.FlagSet) {
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

func (opts *CommonFlagOpts) ApplyFlag(fs *pflag.FlagSet) {
	setFlagMediaType(fs, &opts.MediaType)
	setFlagLocal(fs, &opts.Local)
}

type RemoteFlagOpts struct {
	SecureFlagOpts
	CommonFlagOpts
}

func (opts *RemoteFlagOpts) ApplyFlag(fs *pflag.FlagSet) {
	opts.SecureFlagOpts.ApplyFlag(fs)
	opts.CommonFlagOpts.ApplyFlag(fs)
}
