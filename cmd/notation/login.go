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
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/notaryproject/notation-go/log"
	"github.com/notaryproject/notation/internal/auth"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"oras.land/oras-go/v2/registry/remote/credentials"
)

const urlDocHowToAuthenticate = "https://notaryproject.dev/docs/how-to/registry-authentication/"

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
	ctx = opts.LoggingFlagOpts.InitializeLogger(ctx)

	// initialize
	serverAddress := opts.server

	// input username and password by prompt
	reader := bufio.NewReader(os.Stdin)
	var err error
	if opts.Password == "" {
		var isToken bool
		if opts.Username == "" {
			opts.Username, err = readUsernameFromPrompt(reader)
			if err != nil {
				return err
			}
			if opts.Username == "" {
				// the username is empty, the password is used as a token
				isToken = true
			}
		}
		opts.Password, err = readPasswordFromPrompt(reader, isToken)
		if err != nil {
			return err
		}
		if opts.Password == "" {
			if isToken {
				return errors.New("token required")
			}
			return errors.New("password required")
		}
	}
	cred := opts.Credential()

	credsStore, err := auth.NewCredentialsStore()
	if err != nil {
		return fmt.Errorf("failed to get credentials store: %v", err)
	}
	registry, err := getRegistryLoginClient(ctx, &opts.SecureFlagOpts, serverAddress)
	if err != nil {
		return fmt.Errorf("failed to get registry client: %v", err)
	}
	if err := credentials.Login(ctx, credsStore, registry, cred); err != nil {
		registryName := registry.Reference.Registry
		if !errors.Is(err, credentials.ErrPlaintextPutDisabled) {
			return fmt.Errorf("failed to log in to %s: %v", registryName, err)
		}

		// ErrPlaintextPutDisabled returned by Login() indicates that the
		// credential is validated but is not saved because there is no native
		// credentials store available
		credKeyName := credentials.ServerAddressFromRegistry(registryName)
		if savedCred, err := credsStore.Get(ctx, credKeyName); err != nil || savedCred != cred {
			if err != nil {
				// if we fail to get the saved credential, log a warning
				// but do not throw the GET error, as the error could be
				// confusing to users
				logger := log.GetLogger(ctx)
				logger.Warnf("Failed to get the existing credentials for %s: %v", registryName, err)
			}
			return fmt.Errorf("failed to log in to %s: the credential could not be saved because a credentials store is required to securely store the password. See %s",
				registryName, urlDocHowToAuthenticate)
		}

		// the credential already exists but is in plaintext, ignore the saving error
		fmt.Fprintf(os.Stderr, "Warning: The credentials store is not set up. It is recommended to configure the credentials store to securely store your credentials. See %s.\n", urlDocHowToAuthenticate)
		fmt.Println("Authenticated with existing credentials")
	}

	fmt.Println("Login Succeeded")
	return nil
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

func readPasswordFromPrompt(reader *bufio.Reader, isToken bool) (string, error) {
	var passwordType string
	if isToken {
		passwordType = "token"
		fmt.Print("Token: ")
	} else {
		passwordType = "password"
		fmt.Print("Password: ")
	}

	if term.IsTerminal(int(os.Stdin.Fd())) {
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", fmt.Errorf("error reading %s: %w", passwordType, err)
		}
		fmt.Println()
		return string(bytePassword), nil
	} else {
		password, err := readLine(reader)
		if err != nil {
			return "", fmt.Errorf("error reading %s: %w", passwordType, err)
		}
		fmt.Println()
		return password, nil
	}
}
