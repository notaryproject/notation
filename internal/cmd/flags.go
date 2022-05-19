// Package cmd contains common flags and routines for all CLIs.
package cmd

import (
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

	FlagPluginConfig = &cli.StringSliceFlag{
		Name:    "pluginConfig",
		Aliases: []string{"pc"},
		Usage:   "list of comma-separated {key}={value} pairs that are passed as is to the plugin, refer plugin documentation to set appropriate values",
	}
)

func ParseFlagPluginConfig(pluginConfigSlice []string) (map[string]string, error) {
	if len(pluginConfigSlice) == 0 {
		return nil, nil
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
