package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/notaryproject/notation-go"
	notationRegistry "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

type signOpts struct {
	cmd.LoggingFlagOpts
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
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SignerFlagOpts.ApplyFlags(command.Flags())
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	cmd.SetPflagExpiry(command.Flags(), &opts.expiry)
	cmd.SetPflagPluginConfig(command.Flags(), &opts.pluginConfig)
	return command
}

func runSign(command *cobra.Command, cmdOpts *signOpts) error {
	// set log level
	ctx, _ := cmdOpts.LoggingFlagOpts.SetLoggerLevel(command.Context())

	// initialize
	signer, err := cmd.GetSigner(&cmdOpts.SignerFlagOpts)
	if err != nil {
		return err
	}

	// core process
	desc, opts, err := prepareSigningContent(ctx, cmdOpts)
	if err != nil {
		return err
	}
	remoteRepo, err := getSignatureRepositoryClient(&cmdOpts.SecureFlagOpts, cmdOpts.reference)
	if err != nil {
		return err
	}
	setHttpDebugLog(remoteRepo, cmdOpts.Debug)
	repo := notationRegistry.NewRepository(remoteRepo)
	_, err = notation.Sign(ctx, signer, repo, opts)
	if err != nil {
		return err
	}

	// write out
	fmt.Println(desc.Digest)
	return nil
}

func prepareSigningContent(ctx context.Context, opts *signOpts) (ocispec.Descriptor, notation.SignOptions, error) {
	manifestDesc, err := getManifestDescriptorFromContext(ctx, &opts.SecureFlagOpts, opts.reference)
	if err != nil {
		return ocispec.Descriptor{}, notation.SignOptions{}, err
	}
	mediaType, err := envelope.GetEnvelopeMediaType(opts.SignerFlagOpts.SignatureFormat)
	if err != nil {
		return ocispec.Descriptor{}, notation.SignOptions{}, err
	}
	pluginConfig, err := cmd.ParseFlagPluginConfig(opts.pluginConfig)
	if err != nil {
		return ocispec.Descriptor{}, notation.SignOptions{}, err
	}

	signOpts := notation.SignOptions{
		ArtifactReference:  opts.reference,
		SignatureMediaType: mediaType,
		ExpiryDuration:     opts.expiry,
		PluginConfig:       pluginConfig,
	}

	return manifestDesc, signOpts, nil
}
