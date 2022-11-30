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
)

type tagReference struct {
	isTag bool
	tag   string
}

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
	desc, opts, tagReference, err := prepareSigningContent(command.Context(), cmdOpts)
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
	if tagReference.isTag {
		fmt.Println("Warning: Always sign the artifact using digest(`@sha256:...`) rather than a tag(`:latest`) because tags are mutable and a tag reference can point to a different artifact than the one signed.")
		fmt.Printf("Resolving artifact tag %q to digest %q before signing.\n", tagReference.tag, desc.Digest)
	}
	fmt.Println("Successfully signed", desc.Digest)
	return nil
}

func prepareSigningContent(ctx context.Context, opts *signOpts) (ocispec.Descriptor, notation.SignOptions, tagReference, error) {
	var tagRef tagReference
	isTag := !isDigestReference(opts.reference)
	manifestDesc, ref, err := getManifestDescriptorFromContext(ctx, &opts.SecureFlagOpts, opts.reference)
	if err != nil {
		return ocispec.Descriptor{}, notation.SignOptions{}, tagReference{}, err
	}
	mediaType, err := envelope.GetEnvelopeMediaType(opts.SignerFlagOpts.SignatureFormat)
	if err != nil {
		return ocispec.Descriptor{}, notation.SignOptions{}, tagReference{}, err
	}
	pluginConfig, err := cmd.ParseFlagPluginConfig(opts.pluginConfig)
	if err != nil {
		return ocispec.Descriptor{}, notation.SignOptions{}, tagReference{}, err
	}
	if isTag {
		tagRef = tagReference{
			isTag: isTag,
			tag:   ref.Reference, // ref.Reference is a tag reference
		}
	}
	return manifestDesc, notation.SignOptions{
		ArtifactReference:  opts.reference,
		SignatureMediaType: mediaType,
		Expiry:             cmd.GetExpiry(opts.expiry),
		PluginConfig:       pluginConfig,
	}, tagRef, nil
}
