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
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
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
	ociLayout         bool
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

Example - Sign an OCI artifact referenced in an OCI layout
  notation sign --local-content "<oci_layout_path>@<digest>"

Example - Sign an OCI artifact identified by a tag and referenced in an OCI layout
  notation sign --local-content "<oci_layout_path>:<tag>"

Example - [Experimental] Sign an OCI artifact and use OCI artifact manifest to store the signature:
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
	opts.SignerFlagOpts.ApplyFlagsToCommand(command)
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	cmd.SetPflagExpiry(command.Flags(), &opts.expiry)
	cmd.SetPflagPluginConfig(command.Flags(), &opts.pluginConfig)
	command.Flags().StringVar(&opts.signatureManifest, "signature-manifest", signatureManifestImage, "[Experimental] manifest type for signature. options: \"image\", \"artifact\"")
	cmd.SetPflagUserMetadata(command.Flags(), &opts.userMetadata, cmd.PflagUserMetadataSignUsage)
	command.Flags().BoolVar(&opts.ociLayout, "oci-layout", false, "sign local artifact")
	return command
}

func runSign(command *cobra.Command, cmdOpts *signOpts) error {
	// set log level
	ctx := cmdOpts.LoggingFlagOpts.SetLoggerLevel(command.Context())

	// initialize
	signer, err := cmd.GetSigner(ctx, &cmdOpts.SignerFlagOpts)
	if err != nil {
		return err
	}
	ociImageManifest := cmdOpts.signatureManifest == signatureManifestImage
	var inputType inputType = remoteRegistry
	if cmdOpts.ociLayout {
		inputType = ociLayout
	}
	sigRepo, err := getRepositoryForSign(ctx, inputType, cmdOpts.reference, &cmdOpts.SecureFlagOpts, ociImageManifest)
	if err != nil {
		return err
	}
	signOpts, err := prepareSigningOpts(ctx, cmdOpts, sigRepo)
	if err != nil {
		return err
	}
	_, fullRef, err := resolveReference(ctx, inputType, cmdOpts.reference, sigRepo, func(ref string, manifestDesc ocispec.Descriptor) {
		fmt.Fprintf(os.Stderr, "Warning: Always sign the artifact using digest(@sha256:...) rather than a tag(:%s) because tags are mutable and a tag reference can point to a different artifact than the one signed.\n", ref)
	})
	if err != nil {
		return err
	}
	signOpts.ArtifactReference = fullRef

	// core process
	_, err = notation.Sign(ctx, signer, sigRepo, signOpts)
	if err != nil {
		var errorPushSignatureFailed notation.ErrorPushSignatureFailed
		if errors.As(err, &errorPushSignatureFailed) && !ociImageManifest {
			return fmt.Errorf("%v. Possible reason: OCI artifact manifest is not supported. Try removing the flag `--signature-manifest artifact` to store signatures using OCI image manifest", err)
		}
		return err
	}
	fmt.Println("Successfully signed", fullRef)
	return nil
}

// func signRemote(ctx context.Context, cmdOpts *signOpts, signer notation.Signer, ociImageManifest bool) error {
// 	sigRepo, err := getRemoteRepositoryForSign(ctx, &cmdOpts.SecureFlagOpts, cmdOpts.reference, ociImageManifest)
// 	if err != nil {
// 		return err
// 	}
// 	opts, err := prepareSigningOpts(ctx, cmdOpts, sigRepo)
// 	if err != nil {
// 		return err
// 	}
// 	ref, err := resolveReference(ctx, &cmdOpts.SecureFlagOpts, cmdOpts.reference, sigRepo, func(ref registry.Reference, manifestDesc ocispec.Descriptor) {
// 		fmt.Fprintf(os.Stderr, "Warning: Always sign the artifact using digest(@sha256:...) rather than a tag(:%s) because tags are mutable and a tag reference can point to a different artifact than the one signed.\n", ref.Reference)
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	opts.ArtifactReference = ref.String()

// 	// core process
// 	_, err = notation.Sign(ctx, signer, sigRepo, opts)
// 	if err != nil {
// 		var errorPushSignatureFailed notation.ErrorPushSignatureFailed
// 		if errors.As(err, &errorPushSignatureFailed) && !ociImageManifest {
// 			return fmt.Errorf("%v. Possible reason: target registry does not support OCI artifact manifest. Try removing the flag `--signature-manifest artifact` to store signatures using OCI image manifest", err)
// 		}
// 		return err
// 	}
// 	fmt.Println("Successfully signed", ref)
// 	return nil
// }

// func signLocal(ctx context.Context, cmdOpts *signOpts, signer notation.Signer, ociImageManifest bool) error {
// 	logger := log.GetLogger(ctx)

// 	layoutPath, layoutReference, err := parseOCILayoutReference(cmdOpts.reference)
// 	if err != nil {
// 		return err
// 	}
// 	sigRepo, err := notationregistry.NewOCIRepository(layoutPath, notationregistry.RepositoryOptions{OCIImageManifest: ociImageManifest})
// 	if err != nil {
// 		return err
// 	}
// 	targetDesc, err := sigRepo.Resolve(ctx, layoutReference)
// 	if err != nil {
// 		return fmt.Errorf("failed to resolve OCI layout reference: %s", err)
// 	}
// 	logger.Infof("OCI layout reference %s resolved to target manifest descriptor: %+v", layoutReference, targetDesc)
// 	if digest.Digest(layoutReference).Validate() != nil {
// 		// layout.reference is a tag
// 		fmt.Fprintf(os.Stderr, "Warning: Always sign the artifact using digest(@sha256:...) rather than a tag(:%s) because tags are mutable and a tag reference can point to a different artifact than the one signed.\n", layoutReference)
// 	}
// 	layoutReference = targetDesc.Digest.String()
// 	signOpts, err := prepareSigningOpts(ctx, cmdOpts, sigRepo)
// 	if err != nil {
// 		return err
// 	}
// 	signOpts.ArtifactReference = localTargetPath(layoutPath, layoutReference)

// 	// core process
// 	targetDesc, err = notation.Sign(ctx, signer, sigRepo, signOpts)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("Successfully signed", layoutPath+"@"+targetDesc.Digest.String())
// 	return nil
// }

func prepareSigningOpts(ctx context.Context, opts *signOpts, sigRepo notationregistry.Repository) (notation.SignOptions, error) {
	mediaType, err := envelope.GetEnvelopeMediaType(opts.SignerFlagOpts.SignatureFormat)
	if err != nil {
		return notation.SignOptions{}, err
	}
	pluginConfig, err := cmd.ParseFlagMap(opts.pluginConfig, cmd.PflagPluginConfig.Name)
	if err != nil {
		return notation.SignOptions{}, err
	}
	userMetadata, err := cmd.ParseFlagMap(opts.userMetadata, cmd.PflagUserMetadata.Name)
	if err != nil {
		return notation.SignOptions{}, err
	}
	signOpts := notation.SignOptions{
		SignerSignOptions: notation.SignerSignOptions{
			SignatureMediaType: mediaType,
			ExpiryDuration:     opts.expiry,
			PluginConfig:       pluginConfig,
		},
		UserMetadata: userMetadata,
	}
	return signOpts, nil
}

func validateSignatureManifest(signatureManifest string) bool {
	return slices.Contains(supportedSignatureManifest, signatureManifest)
}
