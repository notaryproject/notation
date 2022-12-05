package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

type signOpts struct {
	cmd.SignerFlagOpts
	SecureFlagOpts
	expiry       time.Duration
	pluginConfig []string
	reference    string
}

func signCommand(opts *signOpts) *cobra.Command {
	if opts == nil {
		opts = &signOpts{}
	}
	command := &cobra.Command{
		Use:   "sign [flags] <reference>",
		Short: "Sign artifacts",
		Long: `Sign artifacts

Prerequisite: a signing key needs to be configured using the command "notation key".

Example - Sign an OCI artifact using the default signing key, with the default JWS envelope:
  notation sign <registry>/<repository>@<digest>

Example - Sign an OCI artifact using the default signing key, with the COSE envelope:
  notation sign --signature-format cose <registry>/<repository>@<digest> 

Example - Sign an OCI artifact using a specified key
  notation sign --key <key_name> <registry>/<repository>@<digest>

Example - Sign an OCI artifact identified by a tag (Notation will resolve tag to digest)
  notation sign <registry>/<repository>:<tag>

Example - Sign an OCI artifact stored in a registry and specify the signature expiry duration, for example 24 hours
  notation sign --expiry 24h <registry>/<repository>@<digest>
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing reference")
			}
			opts.reference = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSign(cmd, opts)
		},
	}
	opts.SignerFlagOpts.ApplyFlags(command.Flags())
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	cmd.SetPflagExpiry(command.Flags(), &opts.expiry)
	cmd.SetPflagPluginConfig(command.Flags(), &opts.pluginConfig)

	return command
}

func runSign(command *cobra.Command, cmdOpts *signOpts) error {
	// initialize
	signer, err := cmd.GetSigner(&cmdOpts.SignerFlagOpts)
	if err != nil {
		return err
	}

	// core process
	opts, ref, err := prepareSigningContent(command.Context(), cmdOpts)
	if err != nil {
		return err
	}
	sigRepo, err := getSignatureRepository(&cmdOpts.SecureFlagOpts, cmdOpts.reference)
	if err != nil {
		return err
	}
	_, err = notation.Sign(command.Context(), signer, sigRepo, opts)
	if err != nil {
		return err
	}

	// write out
	fmt.Println("Successfully signed", ref)

	return nil
}

func prepareSigningContent(ctx context.Context, opts *signOpts) (notation.SignOptions, registry.Reference, error) {
	manifestDesc, ref, err := getManifestDescriptor(ctx, &opts.SecureFlagOpts, opts.reference)
	if err != nil {
		return notation.SignOptions{}, registry.Reference{}, err
	}
	mediaType, err := envelope.GetEnvelopeMediaType(opts.SignerFlagOpts.SignatureFormat)
	if err != nil {
		return notation.SignOptions{}, registry.Reference{}, err
	}
	pluginConfig, err := cmd.ParseFlagPluginConfig(opts.pluginConfig)
	if err != nil {
		return notation.SignOptions{}, registry.Reference{}, err
	}
	if err := ref.ValidateReferenceAsDigest(); err != nil {
		// reference is not a digest reference
		fmt.Printf("Warning: Always sign the artifact using digest(`@sha256:...`) rather than a tag(`:%s`) because tags are mutable and a tag reference can point to a different artifact than the one signed.\n", ref.Reference)
		fmt.Printf("Resolved artifact tag `%s` to digest `%s` before signing.\n", ref.Reference, manifestDesc.Digest.String())

		// resolve tag to digest reference
		ref.Reference = manifestDesc.Digest.String()
	}

	signOpts := notation.SignOptions{
		ArtifactReference:  ref.String(),
		SignatureMediaType: mediaType,
		ExpiryDuration:     opts.expiry,
		PluginConfig:       pluginConfig,
	}

	return signOpts, ref, nil
}
