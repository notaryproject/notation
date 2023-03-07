package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/notaryproject/notation-go"
	notationregistry "github.com/notaryproject/notation-go/registry"
	notationerrors "github.com/notaryproject/notation/cmd/notation/internal/errors"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/internal/osutil"
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
	signatureOutput   string
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
	command.Flags().BoolVar(&opts.localContent, "local-content", false, "if set, sign local content")
	command.Flags().StringVar(&opts.signatureOutput, "output", "", "output path for generated signature")
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
	if cmdOpts.localContent {
		return signLocal(ctx, cmdOpts, signer, ociImageManifest)
	}
	return signRemote(ctx, cmdOpts, signer, ociImageManifest)
}

func signRemote(ctx context.Context, cmdOpts *signOpts, signer notation.Signer, ociImageManifest bool) error {
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
		if errors.As(err, &errorPushSignatureFailed) && !ociImageManifest {
			return fmt.Errorf("%v. Possible reason: target registry does not support OCI artifact manifest. Try removing the flag `--signature-manifest artifact` to store signatures using OCI image manifest", err)
		}
		return err
	}
	fmt.Println("Successfully signed", ref)
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

func signLocal(ctx context.Context, cmdOpts *signOpts, signer notation.Signer, ociImageManifest bool) error {
	mediaType, err := envelope.GetEnvelopeMediaType(cmdOpts.SignerFlagOpts.SignatureFormat)
	if err != nil {
		return err
	}
	pluginConfig, err := cmd.ParseFlagMap(cmdOpts.pluginConfig, cmd.PflagPluginConfig.Name)
	if err != nil {
		return err
	}
	userMetadata, err := cmd.ParseFlagMap(cmdOpts.userMetadata, cmd.PflagUserMetadata.Name)
	if err != nil {
		return err
	}
	localSignOptions := notation.LocalSignOptions{
		SignatureMediaType: mediaType,
		ExpiryDuration:     cmdOpts.expiry,
		PluginConfig:       pluginConfig,
		UserMetadata:       userMetadata,
	}

	var layout ociLayout
	layout.path, layout.reference, err = parseOCILayoutReference(cmdOpts.reference)
	if err != nil {
		var errorOciLayoutMissingReference notationerrors.ErrorOciLayoutMissingReference
		if errors.As(err, &errorOciLayoutMissingReference) {
			isFile, err := osutil.CheckFile(cmdOpts.reference)
			if err != nil {
				return err
			}
			if isFile {
				// descritpor.json
				return signFromFile(ctx, cmdOpts, signer, localSignOptions)
			}
		}
		return err
	}
	return signFromFolder(ctx, cmdOpts, signer, layout, localSignOptions, ociImageManifest)

	// TODO: oci layout tarball
	// sigRepo, err := ociLayoutTarForSign(layout.path, ociImageManifest)
	// if err != nil {
	// 	var errorOciLayoutTarForSign notationerrors.ErrorOciLayoutTarForSign
	// 	if errors.As(err, &errorOciLayoutTarForSign) {
	// 		// oci layout folder
	// 		return signFromFolder(ctx, cmdOpts, signer, layout, localSignOptions, ociImageManifest)
	// 	}
	// 	return err
	// }
	//return signFromTar(ctx, cmdOpts, sigRepo, signer, layout, localSignOptions)
}

func signFromFile(ctx context.Context, cmdOpts *signOpts, signer notation.Signer, localSignOptions notation.LocalSignOptions) error {
	if cmdOpts.signatureOutput == "" {
		return errors.New("signing a descriptor from file, must specifiy output dir for storing generated signature")
	}
	targetDesc, err := getManifestDescriptorFromFile(cmdOpts.reference)
	if err != nil {
		return err
	}

	// core process
	_, sig, _, err := notation.SignLocalContent(ctx, targetDesc, signer, localSignOptions)
	if err != nil {
		return err
	}
	// write out
	output, err := filepath.Abs(cmdOpts.signatureOutput)
	if err != nil {
		return err
	}
	if err := osutil.WriteFileWithPermission(output, sig, 0600, false); err != nil {
		return fmt.Errorf("failed to write generated signature file: %v", err)
	}
	fmt.Println("wrote signature:", output)
	fmt.Println("Successfully signed", cmdOpts.reference+"@"+targetDesc.Digest.String())
	return nil
}

func signFromFolder(ctx context.Context, cmdOpts *signOpts, signer notation.Signer, layout ociLayout, localSignOptions notation.LocalSignOptions, ociImageManifest bool) error {
	sigRepo, err := ociLayoutFolderAsRepositoryForSign(layout.path, ociImageManifest)
	if err != nil {
		return err
	}
	targetDesc, err := sigRepo.Resolve(ctx, layout.reference)
	if err != nil {
		return fmt.Errorf("failed to resolve OCI layout reference: %w", err)
	}
	// layout.reference is a tag
	if digest.Digest(layout.reference).Validate() != nil {
		fmt.Fprintf(os.Stderr, "Warning: Always sign the artifact using digest(@sha256:...) rather than a tag(:%s) because tags are mutable and a tag reference can point to a different artifact than the one signed.\n", layout.reference)
	}

	// core process
	targetDesc, sig, annotations, err := notation.SignLocalContent(ctx, targetDesc, signer, localSignOptions)
	if err != nil {
		return err
	}
	if cmdOpts.signatureOutput != "" {
		path, err := filepath.Abs(cmdOpts.signatureOutput)
		if err != nil {
			return err
		}
		if err := osutil.WriteFileWithPermission(path, sig, 0600, false); err != nil {
			return fmt.Errorf("failed to write generated signature file: %v", err)
		}
		fmt.Println("wrote signature:", path)
	} else {
		_, signatureManifestDesc, err := sigRepo.PushSignature(ctx, localSignOptions.SignatureMediaType, sig, targetDesc, annotations)
		if err != nil {
			return err
		}
		fmt.Printf("Pushed signature to OCI layout folder with manifest digest %q\n", signatureManifestDesc.Digest)
	}
	fmt.Println("Successfully signed", layout.path+"@"+targetDesc.Digest.String())
	return nil
}

// TODO: sign from tarball
// func signFromTar(ctx context.Context, cmdOpts *signOpts, sigRepo notationregistry.Repository, signer notation.Signer, layout ociLayout, localSignOptions notation.LocalSignOptions) error {
// 	if cmdOpts.signatureOutput == "" {
// 		return errors.New("signing an oci layout tarball, must specifiy output dir for storing generated signature")
// 	}
// 	targetDesc, err := sigRepo.Resolve(ctx, layout.reference)
// 	if err != nil {
// 		return fmt.Errorf("failed to resolve OCI layout reference: %w", err)
// 	}
// 	// layout.reference is a tag
// 	if digest.Digest(layout.reference).Validate() != nil {
// 		fmt.Fprintf(os.Stderr, "Warning: Always sign the artifact using digest(@sha256:...) rather than a tag(:%s) because tags are mutable and a tag reference can point to a different artifact than the one signed.\n", layout.reference)
// 	}

// 	// core process
// 	targetDesc, sig, _, err := notation.SignLocalContent(ctx, targetDesc, signer, localSignOptions)
// 	if err != nil {
// 		return err
// 	}
// 	path, err := filepath.Abs(cmdOpts.signatureOutput)
// 	if err != nil {
// 		return err
// 	}
// 	if err := osutil.WriteFileWithPermission(path, sig, 0600, false); err != nil {
// 		return fmt.Errorf("failed to write generated signature file: %v", err)
// 	}
// 	fmt.Println("wrote signature:", path)
// 	fmt.Println("Successfully signed", layout.path+"@"+targetDesc.Digest.String())
// 	return nil
// }

func validateSignatureManifest(signatureManifest string) bool {
	return slices.Contains(supportedSignatureManifest, signatureManifest)
}
