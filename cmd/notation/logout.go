package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/notaryproject/notation/internal/auth"
	"github.com/notaryproject/notation/internal/cmd"
	credentials "github.com/oras-project/oras-credentials-go"
	"github.com/spf13/cobra"
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
