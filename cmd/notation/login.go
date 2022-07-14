package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/notaryproject/notation/pkg/auth"
	"github.com/urfave/cli/v2"
	orasauth "oras.land/oras-go/v2/registry/remote/auth"
)

var loginCommand = &cli.Command{
	Name:  "login",
	Usage: "Provides credentials for authenticated registry operations",
	UsageText: `notation login [options] [server]
	
Example - Login with provided username and password:
	notation login -u <user> -p <password> registry.example.com

Example - Login using $NOTATION_USERNAME $NOTATION_PASSWORD variables:
	notation login registry.example.com`,
	ArgsUsage: "[server]",
	Flags: []cli.Flag{
		flagUsername,
		flagPassword,
		&cli.BoolFlag{
			Name:  "password-stdin",
			Usage: "Take the password from stdin",
		},
	},
	Before: readPassword,
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

func readPassword(ctx *cli.Context) error {
	if ctx.Bool("password-stdin") {
		password, err := readLine()
		if err != nil {
			return err
		}
		ctx.Set(flagPassword.Name, password)
	}
	return nil
}

func readLine() (string, error) {
	passwordBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	password := strings.TrimSuffix(string(passwordBytes), "\n")
	password = strings.TrimSuffix(password, "\r")
	return password, nil
}
