package main

import (
	"errors"
	"fmt"
	"github.com/notaryproject/notation-go/verification"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

type verifyOpts struct {
	RemoteFlagOpts
	reference string
}

func verifyCommand(opts *verifyOpts) *cobra.Command {
	if opts == nil {
		opts = &verifyOpts{}
	}
	command := &cobra.Command{
		Use:   "verify [reference]",
		Short: "Verifies OCI Artifacts",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing reference")
			}
			opts.reference = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVerify(cmd, opts)
		},
	}
	opts.ApplyFlags(command.Flags())
	return command
}

func runVerify(command *cobra.Command, opts *verifyOpts) error {
	verifier, err := getVerifier(opts)
	if err != nil {
		return err
	}

	if _, err := verifier.Verify(command.Context(), opts.reference); err != nil {
		return err
	} else {
		fmt.Println("successfully verified the reference : " + opts.reference)
	}

	return nil
}

func getVerifier(opts *verifyOpts) (*verification.Verifier, error) {
	reference, err := registry.ParseReference(opts.reference)
	if err != nil {
		return nil, err
	}

	repository, err := getRepositoryClient(&opts.SecureFlagOpts, reference)
	if err != nil {
		return nil, err
	}

	verifier, err := verification.NewVerifier(repository)
	if err != nil {
		return nil, err
	}

	return verifier, nil
}
