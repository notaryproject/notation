// Copyright The Notary Project Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package flag contains common flags for all commands.
package flag

import (
	"fmt"
	"strings"
	"time"

	"github.com/notaryproject/notation/v2/internal/config"
	"github.com/notaryproject/notation/v2/internal/envelope"
	"github.com/spf13/pflag"
)

var (
	PflagKey = &pflag.Flag{
		Name:      "key",
		Shorthand: "k",
		Usage:     "signing key name, for a key previously added to notation's key list. This is mutually exclusive with the --id and --plugin flags",
	}
	SetPflagKey = func(fs *pflag.FlagSet, p *string) {
		fs.StringVarP(p, PflagKey.Name, PflagKey.Shorthand, "", PflagKey.Usage)
	}

	PflagSignatureFormat = &pflag.Flag{
		Name:  "signature-format",
		Usage: "signature envelope format, options: \"jws\", \"cose\"",
	}
	SetPflagSignatureFormat = func(fs *pflag.FlagSet, p *string) {
		config, err := config.LoadConfigOnce()
		if err != nil || config.SignatureFormat == "" {
			fs.StringVar(p, PflagSignatureFormat.Name, envelope.JWS, PflagSignatureFormat.Usage)
			return
		}

		// set signatureFormat from config
		fs.StringVar(p, PflagSignatureFormat.Name, config.SignatureFormat, PflagSignatureFormat.Usage)
	}

	PflagID = &pflag.Flag{
		Name:  "id",
		Usage: "key id (required if --plugin is set). This is mutually exclusive with the --key flag",
	}
	SetPflagID = func(fs *pflag.FlagSet, p *string) {
		fs.StringVar(p, PflagID.Name, "", PflagID.Usage)
	}

	PflagPlugin = &pflag.Flag{
		Name:  "plugin",
		Usage: "signing plugin name (required if --id is set). This is mutually exclusive with the --key flag",
	}
	SetPflagPlugin = func(fs *pflag.FlagSet, p *string) {
		fs.StringVar(p, PflagPlugin.Name, "", PflagPlugin.Usage)
	}

	PflagExpiry = &pflag.Flag{
		Name:      "expiry",
		Shorthand: "e",
		Usage:     "optional expiry that provides a \"best by use\" time for the artifact. The duration is specified in minutes(m) and/or hours(h). For example: 12h, 30m, 3h20m",
	}
	SetPflagExpiry = func(fs *pflag.FlagSet, p *time.Duration) {
		fs.DurationVarP(p, PflagExpiry.Name, PflagExpiry.Shorthand, time.Duration(0), PflagExpiry.Usage)
	}

	PflagReference = &pflag.Flag{
		Name:      "reference",
		Shorthand: "r",
		Usage:     "original reference",
	}
	SetPflagReference = func(fs *pflag.FlagSet, p *string) {
		fs.StringVarP(p, PflagReference.Name, PflagReference.Shorthand, "", PflagReference.Usage)
	}

	PflagPluginConfig = &pflag.Flag{
		Name:  "plugin-config",
		Usage: "{key}={value} pairs that are passed as it is to a plugin, refer plugin's documentation to set appropriate values",
	}
	SetPflagPluginConfig = func(fs *pflag.FlagSet, p *[]string) {
		fs.StringArrayVar(p, PflagPluginConfig.Name, nil, PflagPluginConfig.Usage)
	}

	PflagUserMetadata = &pflag.Flag{
		Name:      "user-metadata",
		Shorthand: "m",
	}
	PflagUserMetadataSignUsage   = "{key}={value} pairs that are added to the signature payload"
	PflagUserMetadataVerifyUsage = "user defined {key}={value} pairs that must be present in the signature for successful verification if provided"
	SetPflagUserMetadata         = func(fs *pflag.FlagSet, p *[]string, usage string) {
		fs.StringArrayVarP(p, PflagUserMetadata.Name, PflagUserMetadata.Shorthand, nil, usage)
	}

	PflagReferrersTag = &pflag.Flag{
		Name: "force-referrers-tag",
	}
	SetPflagReferrersTag = func(fs *pflag.FlagSet, p *bool, usage string) {
		fs.BoolVar(p, PflagReferrersTag.Name, false, usage)
	}

	PflagUsername = &pflag.Flag{
		Name:      "username",
		Shorthand: "u",
		Usage:     "username for registry operations (default to $NOTATION_USERNAME if not specified)",
	}
	SetFlagUsername = func(fs *pflag.FlagSet, p *string) {
		fs.StringVarP(p, PflagUsername.Name, PflagUsername.Shorthand, "", PflagUsername.Usage)
	}

	PflagPassword = &pflag.Flag{
		Name:      "password",
		Shorthand: "p",
		Usage:     "password for registry operations (default to $NOTATION_PASSWORD if not specified)",
	}
	SetFlagPassword = func(fs *pflag.FlagSet, p *string) {
		fs.StringVarP(p, PflagPassword.Name, PflagPassword.Shorthand, "", PflagPassword.Usage)
	}

	PflagInsecureRegistry = &pflag.Flag{
		Name:     "insecure-registry",
		Usage:    "use HTTP protocol while connecting to registries. Should be used only for testing",
		DefValue: "false",
	}
	SetFlagInsecureRegistry = func(fs *pflag.FlagSet, p *bool) {
		fs.BoolVar(p, PflagInsecureRegistry.Name, false, PflagInsecureRegistry.Usage)
	}
)

// KeyValueSlice is a flag with type int
type KeyValueSlice interface {
	Set(value string) error
	String() string
}

func ParseFlagMap(c []string, flagName string) (map[string]string, error) {
	m := make(map[string]string, len(c))
	for _, pair := range c {
		key, val, found := strings.Cut(pair, "=")
		if !found || key == "" || val == "" {
			return nil, fmt.Errorf("could not parse flag %s: key-value pair requires \"=\" as separator", flagName)
		}
		m[key] = val
	}
	return m, nil
}
