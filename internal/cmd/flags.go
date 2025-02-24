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

	"github.com/spf13/pflag"
)

var (
	PflagKey = &pflag.Flag{
		Name:      "key",
		Shorthand: "k",
		Usage:     "signing key name, for a key previously added to notation's key list. This is mutually exclusive with the --id and --plugin flags",
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

	PflagReferrersAPI = &pflag.Flag{
		Name: "allow-referrers-api",
	}
	PflagReferrersUsageFormat = "[Experimental] use the Referrers API to %s signatures, if not supported (returns 404), fallback to the Referrers tag schema"
	SetPflagReferrersAPI      = func(fs *pflag.FlagSet, p *bool, usage string) {
		fs.BoolVar(p, PflagReferrersAPI.Name, false, usage)
		fs.MarkHidden(PflagReferrersAPI.Name)
	}

	PflagReferrersTag = &pflag.Flag{
		Name: "force-referrers-tag",
	}
	SetPflagReferrersTag = func(fs *pflag.FlagSet, p *bool, usage string) {
		fs.BoolVar(p, PflagReferrersTag.Name, true, usage)
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
