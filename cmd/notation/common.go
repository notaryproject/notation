package main

import (
	"os"

	"github.com/spf13/cobra"
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
		Usage:     "username for generic remote access",
	}
	setFlagUserName = func(cmd *cobra.Command) {
		cmd.Flags().StringP(flagUsername.Name, flagUsername.Shorthand, os.Getenv(defaultUsernameEnv), flagUsername.Usage)
	}

	flagPassword = &pflag.Flag{
		Name:      "password",
		Shorthand: "p",
		Usage:     "password for generic remote access",
	}
	setFlagPassword = func(cmd *cobra.Command) {
		cmd.Flags().StringP(flagPassword.Name, flagPassword.Shorthand, os.Getenv(defaultPasswordEnv), flagPassword.Usage)
	}

	flagPlainHTTP = &pflag.Flag{
		Name:  "plain-http",
		Usage: "remote access via plain HTTP",
	}
	setFlagPlainHTTP = func(cmd *cobra.Command) {
		cmd.Flags().Bool(flagPlainHTTP.Name, false, flagPlainHTTP.Usage)
	}

	flagMediaType = &pflag.Flag{
		Name:  "media-type",
		Usage: "specify the media type of the manifest read from file or stdin",
	}
	setFlagMediaType = func(cmd *cobra.Command) {
		cmd.Flags().String(flagMediaType.Name, defaultMediaType, flagMediaType.Usage)
	}

	flagOutput = &pflag.Flag{
		Name:      "output",
		Shorthand: "o",
		Usage:     "write signature to a specific path",
	}
	setFlagOutput = func(cmd *cobra.Command) {
		cmd.Flags().StringP(flagOutput.Name, flagOutput.Shorthand, "", flagOutput.Usage)
	}

	flagLocal = &pflag.Flag{
		Name:      "local",
		Shorthand: "l",
		Usage:     "reference is a local file",
	}
	setFlagLocal = func(cmd *cobra.Command) {
		cmd.Flags().StringP(flagLocal.Name, flagLocal.Shorthand, "", flagLocal.Usage)
	}

	// TODO: only support one shortage
	flagSignature = &pflag.Flag{
		Name:      "signature",
		Shorthand: "s",
		Usage:     "signature files",
	}
	setFlagSignature = func(cmd *cobra.Command) {
		cmd.Flags().StringSliceP(flagSignature.Name, flagSignature.Shorthand, []string{}, flagSignature.Usage)
	}
)
