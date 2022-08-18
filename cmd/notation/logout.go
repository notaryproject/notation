package main

import (
	"errors"

	"github.com/notaryproject/notation/pkg/auth"
	"github.com/spf13/cobra"
)

type logoutOpts struct {
	server string
}

func logoutCommand(opts *logoutOpts) *cobra.Command {
	if opts == nil {
		opts = &logoutOpts{}
	}
	return &cobra.Command{
		Use:   "logout [server]",
		Short: "Log out the specified registry hostname",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no hostname specified")
			}
			opts.server = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogout(cmd, opts)
		},
	}
}

func runLogout(cmd *cobra.Command, opts *logoutOpts) error {
	// initialize
	serverAddress := opts.server
	nativeStore, err := auth.GetCredentialsStore(serverAddress)
	if err != nil {
		return err
	}
	return nativeStore.Erase(serverAddress)
}
