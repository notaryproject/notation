package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/notaryproject/notation/pkg/auth"
	"github.com/spf13/cobra"
	orasauth "oras.land/oras-go/v2/registry/remote/auth"
)

type loginOpts struct {
	SecureFlagOpts
	passwordStdin bool
	server        string
}

func loginCommand(opts *loginOpts) *cobra.Command {
	if opts == nil {
		opts = &loginOpts{}
	}
	command := &cobra.Command{
		Use:   "login [flags] <server>",
		Short: "Login to registry",
		Long: `Log in to an OCI registry

Example - Login with provided username and password:
	notation login -u <user> -p <password> registry.example.com

Example - Login using $NOTATION_USERNAME $NOTATION_PASSWORD variables:
	notation login registry.example.com`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no hostname specified")
			}
			opts.server = args[0]
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := readPassword(opts); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(cmd, opts)
		},
	}
	command.Flags().BoolVar(&opts.passwordStdin, "password-stdin", false, "take the password from stdin")
	opts.ApplyFlags(command.Flags())
	return command
}

func runLogin(cmd *cobra.Command, opts *loginOpts) error {
	// initialize
	serverAddress := opts.server

	if err := validateAuthConfig(cmd.Context(), opts, serverAddress); err != nil {
		return err
	}

	nativeStore, err := auth.GetCredentialsStore(serverAddress)
	if err != nil {
		return fmt.Errorf("could not get the credentials store: %v", err)
	}

	// init creds
	creds := newCredentialFromInput(
		opts.Username,
		opts.Password,
	)
	if err = nativeStore.Store(serverAddress, creds); err != nil {
		return fmt.Errorf("failed to store credentials: %v", err)
	}

	fmt.Println("Login Succeeded")
	return nil
}

func validateAuthConfig(ctx context.Context, opts *loginOpts, serverAddress string) error {
	registry, err := getRegistryClient(&opts.SecureFlagOpts, serverAddress)
	if err != nil {
		return err
	}
	return registry.Ping(ctx)
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

func readPassword(opts *loginOpts) error {
	if opts.passwordStdin {
		password, err := readLine()
		if err != nil {
			return err
		}
		opts.Password = password
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
