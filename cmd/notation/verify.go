package main

import (
	"errors"
	"os"

	"github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation-go/verification"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/ioutil"

	orasregistry "oras.land/oras-go/v2/registry"

	"github.com/spf13/cobra"
)

type verifyOpts struct {
	SecureFlagOpts
	reference string
	config    string
}

func verifyCommand(opts *verifyOpts) *cobra.Command {
	if opts == nil {
		opts = &verifyOpts{}
	}
	command := &cobra.Command{
		Use:   "verify [flags] <reference>",
		Short: "Verifies OCI Artifacts",
		Long: `Verifies OCI Artifacts:
  notation verify [--config <key>=<value>,...] [--username <username>] [--password <password>] <reference>`,
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
	command.Flags().StringVarP(&opts.config, "config", "c", "", "list of comma-separated {key}={value} pairs that are passed as is to the plugin")
	return command
}

func runVerify(command *cobra.Command, opts *verifyOpts) error {
	// resolve the given reference and set the digest.
	ref, err := resolveReference(command, opts)
	if err != nil {
		return err
	}

	// initialize verifier.
	verifier, err := getVerifier(opts, ref)
	if err != nil {
		return err
	}

	// set up verification plugin config.
	configs, err := cmd.ParseFlagPluginConfig(opts.config)
	if err != nil {
		return err
	}

	// core verify process.
	ctx := verification.WithPluginConfig(command.Context(), configs)
	outcomes, err := verifier.Verify(ctx, ref.String())

	// write out.
	return ioutil.PrintVerificationResults(os.Stdout, outcomes, err, ref.Reference)
}

func getVerifier(opts *verifyOpts, ref orasregistry.Reference) (*verification.Verifier, error) {
	authClient, plainHTTP, err := getAuthClient(&opts.SecureFlagOpts, ref)
	if err != nil {
		return nil, err
	}

	repo := registry.NewRepositoryClient(authClient, ref, plainHTTP)

	return verification.NewVerifier(repo)
}

func resolveReference(command *cobra.Command, opts *verifyOpts) (orasregistry.Reference, error) {
	ref, err := orasregistry.ParseReference(opts.reference)
	if err != nil {
		return orasregistry.Reference{}, err
	}

	manifestDesc, err := getManifestDescriptorFromReference(command.Context(), &opts.SecureFlagOpts, opts.reference)
	if err != nil {
		return orasregistry.Reference{}, err
	}

	ref.Reference = manifestDesc.Digest.String()
	return ref, nil
}
