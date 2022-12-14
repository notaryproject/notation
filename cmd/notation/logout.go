package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/notaryproject/notation-go/log"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/pkg/auth"
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
	ctx = opts.LoggingFlagOpts.SetLoggerLevel(ctx)
	logger := log.GetLogger(ctx)

	// initialize
	serverAddress := opts.server
	nativeStore, err := auth.GetCredentialsStore(ctx, serverAddress)
	if err != nil {
		return err
	}
	err = nativeStore.Erase(serverAddress)
	if err != nil {
		return err
	}

	logger.Infoln("Logged out from", serverAddress)
	fmt.Println("Logout Succeeded")
	return nil
}
