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
	"errors"
	"fmt"

	"github.com/notaryproject/notation/internal/auth"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry/remote/credentials"
)

type logoutOpts struct {
	cmd.LoggingFlagOpts
	server string
}

func logoutCommand(opts *logoutOpts) *cobra.Command {
	if opts == nil {
		opts = &logoutOpts{}
	}
	command := &cobra.Command{
		Use:   "logout [flags] <server>",
		Short: "Log out from the logged in registries",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no hostname specified")
			}
			opts.server = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogout(cmd.Context(), opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	return command
}

func runLogout(ctx context.Context, opts *logoutOpts) error {
	// set log level
	ctx = opts.LoggingFlagOpts.InitializeLogger(ctx)
	credsStore, err := auth.NewCredentialsStore()
	if err != nil {
		return fmt.Errorf("failed to get credentials store: %v", err)
	}
	if err := credentials.Logout(ctx, credsStore, opts.server); err != nil {
		return fmt.Errorf("failed to log out of %s: %v", opts.server, err)
	}

	fmt.Println("Logout Succeeded")
	return nil
}
