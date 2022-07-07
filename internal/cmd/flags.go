// Package cmd contains common flags and routines for all CLIs.
package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	FlagKey = &pflag.Flag{
		Name:      "key",
		Shorthand: "k",
		Usage:     "signing key name",
	}
	SetFlagKey = func(cmd *cobra.Command) {
		cmd.Flags().StringP(FlagKey.Name, FlagKey.Shorthand, "", FlagKey.Usage)
	}

	FlagKeyFile = &pflag.Flag{
		Name:  "key-file",
		Usage: "signing key file",
	}
	SetFlagKeyFile = func(cmd *cobra.Command) {
		cmd.Flags().String(FlagKeyFile.Name, "", FlagKeyFile.Usage)
	}

	FlagCertFile = &pflag.Flag{
		Name:  "cert-file",
		Usage: "signing certificate file",
	}
	SetFlagCertFile = func(cmd *cobra.Command) {
		cmd.Flags().String(FlagCertFile.Name, "", FlagCertFile.Usage)
	}

	FlagTimestamp = &pflag.Flag{
		Name:      "timestamp",
		Shorthand: "t",
		Usage:     "timestamp the signed signature via the remote TSA",
	}
	SetFlagTimestamp = func(cmd *cobra.Command) {
		cmd.Flags().StringP(FlagTimestamp.Name, FlagTimestamp.Shorthand, "", FlagTimestamp.Usage)
	}

	FlagExpiry = &pflag.Flag{
		Name:      "expiry",
		Shorthand: "e",
		Usage:     "expire duration",
	}
	SetFlagExpiry = func(cmd *cobra.Command) {
		cmd.Flags().DurationP(FlagExpiry.Name, FlagExpiry.Shorthand, time.Duration(0), FlagExpiry.Usage)
	}

	FlagReference = &pflag.Flag{
		Name:      "reference",
		Shorthand: "r",
		Usage:     "original reference",
	}
	SetFlagReference = func(cmd *cobra.Command) {
		cmd.Flags().StringP(FlagReference.Name, FlagReference.Shorthand, "", FlagReference.Usage)
	}

	// TODO: cobra does not support char shorthand
	FlagPluginConfig = &pflag.Flag{
		Name:      "pluginConfig",
		Shorthand: "c",
		Usage:     "list of comma-separated {key}={value} pairs that are passed as is to the plugin, refer plugin documentation to set appropriate values",
	}
	SetFlagPluginConfig = func(cmd *cobra.Command) {
		cmd.Flags().StringP(FlagPluginConfig.Name, FlagPluginConfig.Shorthand, "", FlagPluginConfig.Usage)
	}
)

// KeyValueSlice is a flag with type int
type KeyValueSlice interface {
	Set(value string) error
	String() string
}

func ParseFlagPluginConfig(cmd *cobra.Command) (map[string]string, error) {
	val, _ := cmd.Flags().GetString(FlagPluginConfig.Name)
	pluginConfig, err := ParseKeyValueListFlag(val)
	if err != nil {
		return nil, fmt.Errorf("could not parse %q as value for flag %s: %s", val, FlagPluginConfig.Name, err)
	}
	return pluginConfig, nil
}

func ParseKeyValueListFlag(val string) (map[string]string, error) {
	if val == "" {
		return nil, nil
	}
	flags, err := splitQuoted(val)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(flags))
	for _, c := range flags {
		c := strings.TrimSpace(c)
		if c == "" {
			return nil, fmt.Errorf("empty entry: %q", c)
		}
		if k, v, ok := strings.Cut(c, "="); ok {
			k := strings.TrimSpace(k)
			v := strings.TrimSpace(v)
			if k == "" || v == "" {
				return nil, errors.New("empty key value")
			}
			if _, exist := m[k]; exist {
				return nil, fmt.Errorf("duplicated key: %q", k)
			}
			m[k] = v
		} else {
			return nil, fmt.Errorf("malformed entry: %q", c)
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
//	[]string{"a", "b:c,d", "ef", "", `g"`}
func splitQuoted(s string) (r []string, err error) {
	var args []string
	arg := make([]rune, len(s))
	escaped := false
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
			quote = r
			continue
		case r == ',':
			args = append(args, string(arg[:i]))
			i = 0
			continue
		}
		arg[i] = r
		i++
	}
	args = append(args, string(arg[:i]))
	if quote != 0 {
		err = errors.New("unclosed quote")
	} else if escaped {
		err = errors.New("unfinished escaping")
	}
	return args, err
}
