package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/pkg/auth"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	orasauth "oras.land/oras-go/v2/registry/remote/auth"
)

type loginOpts struct {
	cmd.LoggingFlagOpts
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
			return runLogin(cmd.Context(), opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	command.Flags().BoolVar(&opts.passwordStdin, "password-stdin", false, "take the password from stdin")
	return command
}

func runLogin(ctx context.Context, opts *loginOpts) error {
	// set log level
	ctx = opts.LoggingFlagOpts.SetLoggerLevel(ctx)

	// initialize
	serverAddress := opts.server

	// input username and password by prompt
	reader := bufio.NewReader(os.Stdin)
	var err error
	if opts.Username == "" {
		opts.Username, err = readUsernameFromPrompt(reader)
		if err != nil {
			return err
		}
	}

	if opts.Password == "" {
		opts.Password, err = readPasswordFromPrompt(reader)
		if err != nil {
			return err
		}
	}

	if err := validateAuthConfig(ctx, opts, serverAddress); err != nil {
		return err
	}

	nativeStore, err := auth.GetCredentialsStore(ctx, serverAddress)
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
	registry, err := getRegistryClient(ctx, &opts.SecureFlagOpts, serverAddress)
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
		password, err := readLine(os.Stdin)
		if err != nil {
			return err
		}
		opts.Password = password
	}
	return nil
}

func readLine(r io.Reader) (string, error) {
	passwordBytes, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	password := strings.TrimSuffix(string(passwordBytes), "\n")
	password = strings.TrimSuffix(password, "\r")
	return password, nil
}

func readUsernameFromPrompt(reader *bufio.Reader) (string, error) {
	fmt.Print("Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading username: %w", err)
	}
	username = strings.TrimSpace(username)
	return username, nil
}

func readPasswordFromPrompt(reader *bufio.Reader) (string, error) {
	fmt.Print("Password: ")
	if term.IsTerminal(int(os.Stdin.Fd())) {
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", fmt.Errorf("error reading password: %w", err)
		}
		fmt.Println()
		return string(bytePassword), nil
	} else {
		password, err := readLine(reader)
		if err != nil {
			return "", fmt.Errorf("error reading password: %w", err)
		}
		fmt.Println()
		return password, nil
	}
}
