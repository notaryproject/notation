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

package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation/v2/cmd/notation/blob"
	"github.com/notaryproject/notation/v2/cmd/notation/cert"
	"github.com/notaryproject/notation/v2/cmd/notation/internal/flag"
	"github.com/notaryproject/notation/v2/cmd/notation/plugin"
	"github.com/notaryproject/notation/v2/cmd/notation/policy"
	"github.com/spf13/cobra"
)

func run() error {
	cmd := &cobra.Command{
		Use:          "notation",
		Short:        "Notation - a tool to sign and verify artifacts",
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// unset registry credentials after read the value from environment
			// to avoid leaking credentials
			os.Unsetenv(flag.EnvironmentUsername)
			os.Unsetenv(flag.EnvironmentPassword)

			// update Notation config directory
			if notationConfig := os.Getenv("NOTATION_CONFIG"); notationConfig != "" {
				dir.UserConfigDir = notationConfig
			}

			// update Notation cache directory
			if notationCache := os.Getenv("NOTATION_CACHE"); notationCache != "" {
				dir.UserCacheDir = notationCache
			}

			// update Notation Libexec directory (for plugins)
			if notationLibexec := os.Getenv("NOTATION_LIBEXEC"); notationLibexec != "" {
				dir.UserLibexecDir = notationLibexec
			}
		},
	}
	cmd.AddCommand(
		blob.Cmd(),
		signCommand(nil),
		verifyCommand(nil),
		listCommand(nil),
		cert.Cmd(),
		policy.Cmd(),
		keyCommand(),
		plugin.Cmd(),
		loginCommand(nil),
		logoutCommand(nil),
		versionCommand(),
		inspectCommand(nil),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	return cmd.ExecuteContext(ctx)
}

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}
