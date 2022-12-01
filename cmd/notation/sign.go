package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
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

Example - Sign a container image using the default signing key, with the default JWS envelope:
  notation sign <registry>/<repository>:<tag>

Example - Sign a container image using the default signing key, with the COSE envelope:
  notation sign --signature-format cose <registry>/<repository>:<tag> 

Example - Sign a container image using the specified key name
  notation sign --key <key_name> <registry>/<repository>:<tag>

Example - Sign a container image using the image digest
  notation sign <registry>/<repository>@<digest>
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
	desc, opts, ref, err := prepareSigningContent(command.Context(), cmdOpts)
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
	fmt.Printf("Successfully signed %s/%s@%s\n", ref.Registry, ref.Repository, desc.Digest)

	return nil
}

func prepareSigningContent(ctx context.Context, opts *signOpts) (ocispec.Descriptor, notation.SignOptions, registry.Reference, error) {
	manifestDesc, ref, err := getManifestDescriptorFromContext(ctx, &opts.SecureFlagOpts, opts.reference)
	if err != nil {
		return ocispec.Descriptor{}, notation.SignOptions{}, registry.Reference{}, err
	}
	mediaType, err := envelope.GetEnvelopeMediaType(opts.SignerFlagOpts.SignatureFormat)
	if err != nil {
		return ocispec.Descriptor{}, notation.SignOptions{}, registry.Reference{}, err
	}
	pluginConfig, err := cmd.ParseFlagPluginConfig(opts.pluginConfig)
	if err != nil {
		return ocispec.Descriptor{}, notation.SignOptions{}, registry.Reference{}, err
	}
	return manifestDesc, notation.SignOptions{
		ArtifactReference:  opts.reference,
		SignatureMediaType: mediaType,
		Expiry:             cmd.GetExpiry(opts.expiry),
		PluginConfig:       pluginConfig,
	}, ref, nil
}
