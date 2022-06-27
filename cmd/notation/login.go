package main

import (
	"errors"
	"fmt"

	"github.com/notaryproject/notation/pkg/auth"
	"github.com/urfave/cli/v2"
	orasauth "oras.land/oras-go/v2/registry/remote/auth"
)

var loginCommand = &cli.Command{
	Name:      "login",
	Usage:     "Log in the specified registry hostname",
	ArgsUsage: "<SERVER>",
	Flags: []cli.Flag{
		flagUsername,
		flagPassword,
		flagPlainHTTP,
	},
	Action: runLogin,
}

func runLogin(ctx *cli.Context) error {
	// initialize
	if !ctx.Args().Present() {
		return errors.New("no hostname specified")
	}
	serverAddress := ctx.Args().First()

	if err := validateAuthConfig(ctx, serverAddress); err != nil {
		return err
	}

	nativeStore, err := auth.GetCredentialsStore(serverAddress)
	if err != nil {
		return fmt.Errorf("could not get the credentials store: %v", err)
	}
	// init creds
	creds := newCredentialFromInput(
		ctx.String(flagUsername.Name),
		ctx.String(flagPassword.Name),
	)
	if err = nativeStore.Store(serverAddress, creds); err != nil {
		return fmt.Errorf("failed to store credentials: %v", err)
	}
	return nil
}

func validateAuthConfig(ctx *cli.Context, serverAddress string) error {
	registry, err := getRegistryClient(ctx, serverAddress)
	if err != nil {
		return err
	}
	return registry.Ping(ctx.Context)
}

func newCredentialFromInput(username, password string) orasauth.Credential {
	c := orasauth.Credential{
		Username: username,
		Password: password,
	}
	if c.Username == "" {
		c.RefreshToken = password
	}
	return c
}
