package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/notaryproject/notation-go"
	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/cmd/notation/internal/experimental"
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

const referrersTagSchemaDeleteError = "failed to delete dangling referrers index"

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
	inputType         inputType
}

func signCommand(opts *signOpts) *cobra.Command {
	if opts == nil {
		opts = &signOpts{
			inputType: inputTypeRegistry, // remote registry by default
		}
	}
	longMessage := `Sign artifacts

Note: a signing key must be specified. This can be done temporarily by specifying a key ID, or a new key can be configured using the command "notation key add"

Example - Sign an OCI artifact using the default signing key, with the default JWS envelope, and use OCI image manifest to store the signature:
  notation sign <registry>/<repository>@<digest>

Example - Sign an OCI artifact using the default signing key, with the COSE envelope:
  notation sign --signature-format cose <registry>/<repository>@<digest> 

Example - Sign an OCI artifact with a specified plugin and signing key stored in KMS 
  notation sign --plugin <plugin_name> --id <remote_key_id> <registry>/<repository>@<digest>

Example - Sign an OCI artifact using a specified key
  notation sign --key <key_name> <registry>/<repository>@<digest>

Example - Sign an OCI artifact identified by a tag (Notation will resolve tag to digest)
  notation sign <registry>/<repository>:<tag>

Example - Sign an OCI artifact stored in a registry and specify the signature expiry duration, for example 24 hours
  notation sign --expiry 24h <registry>/<repository>@<digest>
`
	experimentalExamples := `
Example - [Experimental] Sign an OCI artifact referenced in an OCI layout
  notation sign --oci-layout "<oci_layout_path>@<digest>"

Example - [Experimental] Sign an OCI artifact identified by a tag and referenced in an OCI layout
  notation sign --oci-layout "<oci_layout_path>:<tag>"

Example - [Experimental] Sign an OCI artifact and use OCI artifact manifest to store the signature:
  notation sign --signature-manifest artifact <registry>/<repository>@<digest>
`

	command := &cobra.Command{
		Use:   "sign [flags] <reference>",
		Short: "Sign artifacts",
		Long:  longMessage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing reference")
			}
			opts.reference = args[0]
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.ociLayout {
				opts.inputType = inputTypeOCILayout
			}
			return experimental.CheckFlagsAndWarn(cmd, "signature-manifest", "oci-layout")
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
	command.Flags().BoolVar(&opts.ociLayout, "oci-layout", false, "[Experimental] sign the artifact stored as OCI image layout")
	experimental.HideFlags(command, experimentalExamples, []string{"signature-manifest", "oci-layout"})
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
	sigRepo, err := getRepositoryForSign(ctx, cmdOpts.inputType, cmdOpts.reference, &cmdOpts.SecureFlagOpts, ociImageManifest)
	if err != nil {
		return err
	}
	signOpts, err := prepareSigningOpts(ctx, cmdOpts, sigRepo)
	if err != nil {
		return err
	}
	manifestDesc, resolvedRef, err := resolveReference(ctx, cmdOpts.inputType, cmdOpts.reference, sigRepo, func(ref string, manifestDesc ocispec.Descriptor) {
		fmt.Fprintf(os.Stderr, "Warning: Always sign the artifact using digest(@sha256:...) rather than a tag(:%s) because tags are mutable and a tag reference can point to a different artifact than the one signed.\n", ref)
	})
	if err != nil {
		return err
	}
	signOpts.ArtifactReference = manifestDesc.Digest.String()

	// core process
	_, err = notation.Sign(ctx, signer, sigRepo, signOpts)
	if err != nil {
		var errorPushSignatureFailed notation.ErrorPushSignatureFailed
		if errors.As(err, &errorPushSignatureFailed) {
			if !ociImageManifest {
				return fmt.Errorf("%v. Possible reason: OCI artifact manifest is not supported. Try removing the flag `--signature-manifest artifact` to store signatures using OCI image manifest", err)
			}
			if strings.Contains(err.Error(), referrersTagSchemaDeleteError) {
				fmt.Fprintln(os.Stderr, "Warning: Removal of outdated referrers index from remote registry failed. Garbage collection may be required.")
				// write out
				fmt.Println("Successfully signed", resolvedRef)
				return nil
			}
		}
		return err
	}
	fmt.Println("Successfully signed", resolvedRef)
	return nil
}

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
