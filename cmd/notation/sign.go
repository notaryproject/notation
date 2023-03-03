package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/notaryproject/notation-go"
	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/internal/slices"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

const (
	signatureManifestArtifact = "artifact"
	signatureManifestImage    = "image"
)

var supportedSignatureManifest = []string{signatureManifestArtifact, signatureManifestImage}

type signOpts struct {
	cmd.LoggingFlagOpts
	cmd.SignerFlagOpts
	SecureFlagOpts
	expiry            time.Duration
	pluginConfig      []string
	userMetadata      []string
	reference         string
	signatureManifest string
	localContent      bool
}

type ociLayout struct {
	path      string
	reference string
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

Example - Sign an OCI artifact using the default signing key, with the default JWS envelope, and use OCI image manifest to store the signature:
  notation sign <registry>/<repository>@<digest>

Example - Sign an OCI artifact using the default signing key, with the COSE envelope:
  notation sign --signature-format cose <registry>/<repository>@<digest> 

Example - Sign an OCI artifact using a specified key
  notation sign --key <key_name> <registry>/<repository>@<digest>

Example - Sign an OCI artifact identified by a tag (Notation will resolve tag to digest)
  notation sign <registry>/<repository>:<tag>

Example - Sign an OCI artifact stored in a registry and specify the signature expiry duration, for example 24 hours
  notation sign --expiry 24h <registry>/<repository>@<digest>

Example - Sign an OCI artifact and use OCI artifact manifest to store the signature:
  notation sign --signature-manifest artifact <registry>/<repository>@<digest>
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing reference")
			}
			opts.reference = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// sanity check
			if !validateSignatureManifest(opts.signatureManifest) {
				return fmt.Errorf("signature manifest must be one of the following %v but got %s", supportedSignatureManifest, opts.signatureManifest)
			}
			return runSign(cmd, opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SignerFlagOpts.ApplyFlags(command.Flags())
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	cmd.SetPflagExpiry(command.Flags(), &opts.expiry)
	cmd.SetPflagPluginConfig(command.Flags(), &opts.pluginConfig)
	command.Flags().StringVar(&opts.signatureManifest, "signature-manifest", signatureManifestImage, "manifest type for signature. options: \"image\", \"artifact\"")
	cmd.SetPflagUserMetadata(command.Flags(), &opts.userMetadata, cmd.PflagUserMetadataSignUsage)
	command.Flags().BoolVar(&opts.localContent, "local-content", false, "if set, sign local content")
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
	ociImageManifest := cmdOpts.signatureManifest == signatureManifestImage
	if cmdOpts.localContent {
		var layout ociLayout
		layout.path, layout.reference, err = parseOCILayoutReference(cmdOpts.reference)
		if err != nil {
			return err
		}
		sigRepo, err := ociLayoutFolderAsRepositoryForSign(layout.path, ociImageManifest)
		if err != nil {
			return err
		}
		targetDesc, localSignOpts, err := prepareLocalSigningContent(ctx, cmdOpts, &layout, sigRepo)
		if err != nil {
			return err
		}

		// core process
		targetDesc, sig, annotations, err := notation.SignLocalContent(ctx, targetDesc, signer, localSignOpts)
		if err != nil {
			return err
		}
		_, _, err = sigRepo.PushSignature(ctx, localSignOpts.SignatureMediaType, sig, targetDesc, annotations)
		if err != nil {
			return err
		}
		fmt.Println("Successfully signed", layout.path+"@"+targetDesc.Digest.String())
	} else {
		sigRepo, err := getSignatureRepositoryForSign(ctx, &cmdOpts.SecureFlagOpts, cmdOpts.reference, ociImageManifest)
		if err != nil {
			return err
		}
		opts, ref, err := prepareRemoteSigningContent(ctx, cmdOpts, sigRepo)
		if err != nil {
			return err
		}

		// core process
		_, err = notation.Sign(ctx, signer, sigRepo, opts)
		if err != nil {
			var errorPushSignatureFailed notation.ErrorPushSignatureFailed
			if errors.As(err, &errorPushSignatureFailed) {
				return fmt.Errorf("%v. Target registry does not seem to support OCI artifact manifest. Try the flag `--signature-manifest image` to store signatures using OCI image manifest for backwards compatibility", err)
			}
			return err
		}
		fmt.Println("Successfully signed", ref)
	}
	return nil
}

func prepareRemoteSigningContent(ctx context.Context, opts *signOpts, sigRepo notationregistry.Repository) (notation.RemoteSignOptions, registry.Reference, error) {
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
	ref, err := resolveReference(ctx, &opts.SecureFlagOpts, opts.reference, sigRepo, func(ref registry.Reference, manifestDesc ocispec.Descriptor) {
		fmt.Fprintf(os.Stderr, "Warning: Always sign the artifact using digest(@sha256:...) rather than a tag(:%s) because tags are mutable and a tag reference can point to a different artifact than the one signed.\n", ref.Reference)
	})
	if err != nil {
		return notation.RemoteSignOptions{}, registry.Reference{}, err
	}
	signOpts := notation.RemoteSignOptions{
		SignOptions: notation.SignOptions{
			ArtifactReference:  ref.String(),
			SignatureMediaType: mediaType,
			ExpiryDuration:     opts.expiry,
			PluginConfig:       pluginConfig,
		},
		UserMetadata: userMetadata,
	}
	return signOpts, ref, nil
}

func prepareLocalSigningContent(ctx context.Context, opts *signOpts, layout *ociLayout, sigRepo notationregistry.Repository) (ocispec.Descriptor, notation.LocalSignOptions, error) {
	mediaType, err := envelope.GetEnvelopeMediaType(opts.SignerFlagOpts.SignatureFormat)
	if err != nil {
		return ocispec.Descriptor{}, notation.LocalSignOptions{}, err
	}
	pluginConfig, err := cmd.ParseFlagMap(opts.pluginConfig, cmd.PflagPluginConfig.Name)
	if err != nil {
		return ocispec.Descriptor{}, notation.LocalSignOptions{}, err
	}
	userMetadata, err := cmd.ParseFlagMap(opts.userMetadata, cmd.PflagUserMetadata.Name)
	if err != nil {
		return ocispec.Descriptor{}, notation.LocalSignOptions{}, err
	}
	targetDesc, err := sigRepo.Resolve(ctx, layout.reference)
	if err != nil {
		return ocispec.Descriptor{}, notation.LocalSignOptions{}, err
	}
	// layout.reference is a tag
	if digest.Digest(layout.reference).Validate() != nil {
		fmt.Fprintf(os.Stderr, "Warning: Always sign the artifact using digest(@sha256:...) rather than a tag(:%s) because tags are mutable and a tag reference can point to a different artifact than the one signed.\n", layout.reference)
	}

	localSignOptions := notation.LocalSignOptions{
		SignatureMediaType: mediaType,
		ExpiryDuration:     opts.expiry,
		PluginConfig:       pluginConfig,
		UserMetadata:       userMetadata,
	}
	return targetDesc, localSignOptions, nil
}

func validateSignatureManifest(signatureManifest string) bool {
	return slices.Contains(supportedSignatureManifest, signatureManifest)
}
