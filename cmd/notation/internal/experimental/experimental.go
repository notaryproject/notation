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

// IsDisabled determines whether experimental features are disabled.
func IsDisabled() bool {
	return os.Getenv(envName) != enabled
}

// CheckCommandAndWarn checks whether an experimental command can be run.
func CheckCommandAndWarn(cmd *cobra.Command, _ []string) error {
	return CheckAndWarn(func() (string, bool) {
		return fmt.Sprintf("%q", cmd.CommandPath()), true
	})
}

// CheckFlagsAndWarn checks whether experimental flags can be run.
func CheckFlagsAndWarn(cmd *cobra.Command, flags ...string) error {
	return CheckAndWarn(func() (string, bool) {
		var changedFlags []string
		flagSet := cmd.Flags()
		for _, flag := range flags {
			if flagSet.Changed(flag) {
				changedFlags = append(changedFlags, "--"+flag)
			}
		}
		if len(changedFlags) == 0 {
			// no experimental flag used
			return "", false
		}
		return fmt.Sprintf("flag(s) %s in %q", strings.Join(changedFlags, ","), cmd.CommandPath()), true
	})
}

// CheckAndWarn checks whether a feature can be used.
func CheckAndWarn(doCheck func() (feature string, isExperimental bool)) error {
	feature, isExperimental := doCheck()
	if isExperimental {
		if IsDisabled() {
			// feature is experimental and disabled
			return fmt.Errorf("%s is experimental and not enabled by default. To use, please set %s=%s environment variable", feature, envName, enabled)
		}
		return warn()
	}
	return nil
}

// warn prints a warning message for using the experimental feature.
func warn() error {
	_, err := fmt.Fprintf(os.Stderr, "Warning: This feature is experimental and may not be fully tested or completed and may be deprecated. Report any issues to \"https://github/notaryproject/notation\"\n")
	return err
}

// HideFlags hides experimental flags when NOTATION_EXPERIMENTAL is disabled
// and updates the command's long message accordingly when NOTATION_EXPERIMENTAL
// is enabled.
func HideFlags(cmd *cobra.Command, experimentalExamples string, flags []string) {
	if IsDisabled() {
		flagsSet := cmd.Flags()
		for _, flag := range flags {
			flagsSet.MarkHidden(flag)
		}
	} else if experimentalExamples != "" {
		cmd.Long += experimentalExamples
	}
}
