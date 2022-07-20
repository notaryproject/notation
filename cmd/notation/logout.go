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
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.server = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogout(cmd, opts)
		},
	}
}

func runLogout(cmd *cobra.Command, opts *logoutOpts) error {
	// initialize
	if opts.server == "" {
		return errors.New("no hostname specified")
	}
	serverAddress := opts.server
	nativeStore, err := auth.GetCredentialsStore(serverAddress)
	if err != nil {
		return err
	}
	return nativeStore.Erase(serverAddress)
}
