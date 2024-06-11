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

// Package cmd contains common flags and routines for all CLIs.
package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/pkg/configutil"
	"github.com/spf13/pflag"
)

const (
	OutputPlaintext = "text"
	OutputJSON      = "json"
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
		defaultSignatureFormat := envelope.JWS
		// load config to get signatureFormat
		config, err := configutil.LoadConfigOnce()
		if err == nil && config.SignatureFormat != "" {
			defaultSignatureFormat = config.SignatureFormat
		}

		fs.StringVar(p, PflagSignatureFormat.Name, defaultSignatureFormat, PflagSignatureFormat.Usage)
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

	PflagReferrersAPI = &pflag.Flag{
		Name: "allow-referrers-api",
	}
	PflagReferrersUsageFormat = "[Experimental] use the Referrers API to %s signatures, if not supported (returns 404), fallback to the Referrers tag schema"
	SetPflagReferrersAPI      = func(fs *pflag.FlagSet, p *bool, usage string) {
		fs.BoolVar(p, PflagReferrersAPI.Name, false, usage)
	}

	PflagOutput = &pflag.Flag{
		Name:      "output",
		Shorthand: "o",
	}
	PflagOutputUsage = fmt.Sprintf("output format, options: '%s', '%s'", OutputJSON, OutputPlaintext)
	SetPflagOutput   = func(fs *pflag.FlagSet, p *string, usage string) {
		fs.StringVarP(p, PflagOutput.Name, PflagOutput.Shorthand, OutputPlaintext, usage)
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
