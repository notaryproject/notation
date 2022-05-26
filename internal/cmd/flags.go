// Package cmd contains common flags and routines for all CLIs.
package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

var (
	FlagKey = &cli.StringFlag{
		Name:    "key",
		Aliases: []string{"k"},
		Usage:   "signing key name",
	}

	FlagKeyFile = &cli.StringFlag{
		Name:      "key-file",
		Usage:     "signing key file",
		TakesFile: true,
	}

	FlagCertFile = &cli.StringFlag{
		Name:      "cert-file",
		Usage:     "signing certificate file",
		TakesFile: true,
	}

	FlagTimestamp = &cli.StringFlag{
		Name:    "timestamp",
		Aliases: []string{"t"},
		Usage:   "timestamp the signed signature via the remote TSA",
	}

	FlagExpiry = &cli.DurationFlag{
		Name:    "expiry",
		Aliases: []string{"e"},
		Usage:   "expire duration",
	}

	FlagReference = &cli.StringFlag{
		Name:    "reference",
		Aliases: []string{"r"},
		Usage:   "original reference",
	}

	FlagPluginConfig = &cli.StringFlag{
		Name:    "pluginConfig",
		Aliases: []string{"pc"},
		Usage:   "list of comma-separated {key}={value} pairs that are passed as is to the plugin, refer plugin documentation to set appropriate values",
	}
)

func ParseFlagPluginConfig(v string) (map[string]string, error) {
	if v == "" {
		return nil, nil
	}
	pluginConfigSlice, err := splitQuoted(v)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(pluginConfigSlice))
	for _, c := range pluginConfigSlice {
		if k, v, ok := strings.Cut(c, "="); ok {
			if _, exist := m[k]; exist {
				return nil, fmt.Errorf("duplicated --pluginConfig entry %s", k)
			}
			m[k] = v
		} else {
			return nil, fmt.Errorf("malformed --pluginConfig entry %q", c)
		}
	}
	return m, nil
}

// splitQuoted splits the string s around each instance of one or more consecutive
// comma characters while taking into account quotes and escaping, and
// returns an array of substrings of s or an empty list if s is empty.
// Single quotes and double quotes are recognized to prevent splitting within the
// quoted region, and are removed from the resulting substrings. If a quote in s
// isn't closed err will be set and r will have the unclosed argument as the
// last element. The backslash is used for escaping.
//
// For example, the following string:
//
//	`a,b:"c,d",'e''f',,"g\""`
//
// Would be parsed as:
//
//	[]string{"a", "b:c,d", "ef", `g"`}
func splitQuoted(s string) (r []string, err error) {
	var args []string
	arg := make([]rune, len(s))
	escaped := false
	quoted := false
	quote := '\x00'
	i := 0
	for _, r := range s {
		switch {
		case escaped:
			escaped = false
		case r == '\\':
			escaped = true
			continue
		case quote != 0:
			if r == quote {
				quote = 0
				continue
			}
		case r == '"' || r == '\'':
			quoted = true
			quote = r
			continue
		case r == ',':
			if quoted || i > 0 {
				quoted = false
				args = append(args, string(arg[:i]))
				i = 0
			}
			continue
		}
		arg[i] = r
		i++
	}
	if quoted || i > 0 {
		args = append(args, string(arg[:i]))
	}
	if quote != 0 {
		err = errors.New("unclosed quote")
	} else if escaped {
		err = errors.New("unfinished escaping")
	}
	return args, err
}
