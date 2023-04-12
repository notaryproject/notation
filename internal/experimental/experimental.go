package experimental

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const (
	envName = "NOTATION_EXPERIMENTAL"
	enabled = "1"
)

// IsDisabled determines whether an experimental feature is disabled.
func IsDisabled() bool {
	return os.Getenv(envName) != enabled
}

// Error returns an error for a disabled experimental feature.
func Error(description string) error {
	return fmt.Errorf("%s been marked as experimental and not enabled by default. To use it, please set %s=%s in your environment", description, envName, enabled)
}

// CheckCommandAndWarn checks whether an experimental command can be run.
func CheckCommandAndWarn(cmd *cobra.Command, args []string) error {
	if err := Check(func() (string, bool) {
		return fmt.Sprintf("%q", cmd.CommandPath()), true
	}); err != nil {
		return err
	}
	return Warn()
}

// CheckFlagsAndWarn checks whether experimental flags can be run.
func CheckFlagsAndWarn(cmd *cobra.Command, flags ...string) error {
	if err := Check(func() (string, bool) {
		var changedFlags []string
		flagSet := cmd.Flags()
		for _, flag := range flags {
			flagSet.MarkHidden(flag)
			if flagSet.Changed(flag) {
				changedFlags = append(changedFlags, "--"+flag)
			}
		}
		if len(changedFlags) == 0 {
			// no experimental flag used
			return "", false
		}
		return fmt.Sprintf("flag(s) %s in %q", strings.Join(changedFlags, ","), cmd.CommandPath()), true
	}); err != nil {
		return err
	}
	return Warn()
}

// Check checks whether a feature can be used.
func Check(doCheck func() (feature string, isExperimental bool)) error {
	if IsDisabled() {
		feature, isExperimental := doCheck()
		if isExperimental {
			// feature is experimental and disabled
			return Error(feature)
		}
		return nil
	}
	return nil
}

// Warn prints a warning message for using the experimental feature.
func Warn() error {
	_, err := fmt.Fprintf(os.Stderr, "Caution: This feature is experimental and may not be fully tested or completed and may be deprecated. Report any issues to \"https://github/notaryproject/notation\"\n")
	return err
}
