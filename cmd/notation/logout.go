package main

import (
	"errors"

	"github.com/notaryproject/notation/pkg/auth"
	"github.com/urfave/cli/v2"
)

var logoutCommand = &cli.Command{
	Name:      "logout",
	Usage:     "Log out the specified registry hostname",
	ArgsUsage: "[server]",
	Action:    runLogout,
}

func runLogout(ctx *cli.Context) error {
	// initialize
	if !ctx.Args().Present() {
		return errors.New("no hostname specified")
	}
	serverAddress := ctx.Args().First()
	nativeStore, err := auth.GetCredentialsStore(serverAddress)
	if err != nil {
		return err
	}
	return nativeStore.Erase(serverAddress)
}
