package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/notaryproject/notation-go"
	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

type signOpts struct {
	cmd.LoggingFlagOpts
	cmd.SignerFlagOpts
	SecureFlagOpts
	expiry       time.Duration
	pluginConfig []string
	userMetadata []string
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
	cmd.SetPflagUserMetadata(command.Flags(), &opts.userMetadata, cmd.PflagUserMetadataSignUsage)
	return command
}

func runSign(command *cobra.Command, cmdOpts *signOpts) error {
	// set log level
	ctx := cmdOpts.LoggingFlagOpts.SetLoggerLevel(command.Context())

	// initialize
	signer, err := cmd.GetSigner(&cmdOpts.SignerFlagOpts)
	if err != nil {
		return err
	}
	sigRepo, err := getSignatureRepository(ctx, &cmdOpts.SecureFlagOpts, cmdOpts.reference)
	if err != nil {
		return err
	}
	opts, ref, err := prepareSigningContent(ctx, cmdOpts, sigRepo)
	if err != nil {
		return err
	}

	// core process
	_, err = notation.Sign(ctx, signer, sigRepo, opts)
	if err != nil {
		return err
	}

	// write out
	fmt.Println("Successfully signed", ref)
	return nil
}

func prepareSigningContent(ctx context.Context, opts *signOpts, sigRepo notationregistry.Repository) (notation.RemoteSignOptions, registry.Reference, error) {
	ref, err := resolveReference(ctx, &opts.SecureFlagOpts, opts.reference, sigRepo, func(ref registry.Reference, manifestDesc ocispec.Descriptor) {
		fmt.Printf("Warning: Always sign the artifact using digest(`@sha256:...`) rather than a tag(`:%s`) because tags are mutable and a tag reference can point to a different artifact than the one signed.\n", ref.Reference)
		fmt.Printf("Resolved artifact tag `%s` to digest `%s` before signing.\n", ref.Reference, manifestDesc.Digest.String())
	})
	if err != nil {
		return notation.RemoteSignOptions{}, registry.Reference{}, err
	}

	mediaType, err := envelope.GetEnvelopeMediaType(opts.SignerFlagOpts.SignatureFormat)
	if err != nil {
		return notation.RemoteSignOptions{}, registry.Reference{}, err
	}
	pluginConfig, err := cmd.ParseFlagMap(opts.pluginConfig, cmd.PflagPluginConfig.Name)
	if err != nil {
		return notation.RemoteSignOptions{}, registry.Reference{}, err
	}
	userMetadata, err := cmd.ParseFlagMap(opts.userMetadata, cmd.PflagUserMetadata.Name)
	if err != nil {
		return notation.RemoteSignOptions{}, registry.Reference{}, err
	}

	signOpts := notation.RemoteSignOptions{
	    ArtifactReference: ref.String(),
	    SignatureMediaType: mediaType,
	    ExpiryDuration: opts.expiry,
	    PluginConfig: pluginConfig,
	    UserMetadata: userMetadata,
	}
	return signOpts, ref, nil
}
