// Copyright The Notary Project Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/notaryproject/notation-core-go/revocation/purpose"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/log"
	"github.com/notaryproject/notation/v2/cmd/notation/internal/experimental"
	"github.com/notaryproject/notation/v2/cmd/notation/internal/flag"
	"github.com/notaryproject/notation/v2/cmd/notation/internal/sign"
	"github.com/notaryproject/notation/v2/internal/dmverity"
	"github.com/notaryproject/notation/v2/internal/envelope"
	"github.com/notaryproject/notation/v2/internal/httputil"
	"github.com/notaryproject/notation/v2/internal/registryutil"
	clirev "github.com/notaryproject/notation/v2/internal/revocation"
	nx509 "github.com/notaryproject/notation/v2/internal/x509"
	"github.com/notaryproject/tspclient-go"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

// timestampingTimeout is the timeout when requesting timestamp countersignature
// from a TSA
const timestampingTimeout = 15 * time.Second

type signOpts struct {
	flag.LoggingFlagOpts
	flag.SignerFlagOpts
	flag.SecureFlagOpts
	expiry                 time.Duration
	pluginConfig           []string
	userMetadata           []string
	reference              string
	forceReferrersTag      bool
	ociLayout              bool
	inputType              inputType
	tsaServerURL           string
	tsaRootCertificatePath string
	dmVerity               bool
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

Example - Sign an OCI artifact and store signature using the Referrers API. If it's not supported, fallback to the Referrers tag schema
  notation sign --force-referrers-tag=false <registry>/<repository>@<digest>

Example - Sign an OCI artifact with timestamping:
  notation sign --timestamp-url <TSA_url> --timestamp-root-cert <TSA_root_certificate_filepath> <registry>/<repository>@<digest> 

Example - Sign an OCI artifact with dm-verity layer signatures (PKCS#7) and manifest signature (JWS/COSE):
  notation sign --dm-verity --id <key_id> <registry>/<repository>@<digest>
`
	experimentalExamples := `
Example - [Experimental] Sign an OCI artifact referenced in an OCI layout
  notation sign --oci-layout "<oci_layout_path>@<digest>"

Example - [Experimental] Sign an OCI artifact identified by a tag and referenced in an OCI layout
  notation sign --oci-layout "<oci_layout_path>:<tag>"

Example - [Experimental] Sign an OCI artifact with dm-verity layer signatures (PKCS#7) and manifest signature (COSE):
  notation sign --dm-verity --signature-format cose --id <key_id> <registry>/<repository>@<digest>
`

	command := &cobra.Command{
		Use:   "sign [flags] <reference>",
		Short: "Sign artifacts",
		Long:  longMessage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing reference to the artifact: use `notation sign --help` to see what parameters are required")
			}
			opts.reference = args[0]
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.ociLayout {
				opts.inputType = inputTypeOCILayout
			}
			return experimental.CheckFlagsAndWarn(cmd, "oci-layout", "dm-verity")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// timestamping
			if cmd.Flags().Changed("timestamp-url") {
				if opts.tsaServerURL == "" {
					return errors.New("timestamping: tsa url cannot be empty")
				}
				if opts.tsaRootCertificatePath == "" {
					return errors.New("timestamping: tsa root certificate path cannot be empty")
				}
			}

			// dm-verity mode: layers use PKCS#7, manifest uses specified format (default: JWS)
			if opts.dmVerity {
				// Validation complete - dm-verity mode active
			}

			return runSign(cmd, opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SignerFlagOpts.ApplyFlagsToCommand(command)
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	flag.SetPflagExpiry(command.Flags(), &opts.expiry)
	flag.SetPflagPluginConfig(command.Flags(), &opts.pluginConfig)
	flag.SetPflagUserMetadata(command.Flags(), &opts.userMetadata, flag.PflagUserMetadataSignUsage)
	command.Flags().StringVar(&opts.tsaServerURL, "timestamp-url", "", "RFC 3161 Timestamping Authority (TSA) server URL")
	command.Flags().StringVar(&opts.tsaRootCertificatePath, "timestamp-root-cert", "", "filepath of timestamp authority root certificate")
	flag.SetPflagReferrersTag(command.Flags(), &opts.forceReferrersTag, "force to store signatures using the referrers tag schema")
	command.Flags().BoolVar(&opts.ociLayout, "oci-layout", false, "[Experimental] sign the artifact stored as OCI image layout")
	command.Flags().BoolVar(&opts.dmVerity, "dm-verity", false, `[Experimental] sign image layers with dm-verity for kernel-level integrity verification.
Generates PKCS#7 signatures of dm-verity root hashes for each layer, compatible with
Linux kernel dm-verity and containerd erofs-snapshotter. Requires: mkfs.erofs, veritysetup`)
	command.MarkFlagsMutuallyExclusive("oci-layout", "force-referrers-tag")
	command.MarkFlagsRequiredTogether("timestamp-url", "timestamp-root-cert")
	experimental.HideFlags(command, experimentalExamples, []string{"oci-layout", "dm-verity"})
	return command
}

// fetchImageManifest fetches and parses an OCI manifest using registryutil.BlobFetcher.
func fetchImageManifest(ctx context.Context, secureOpts *flag.SecureFlagOpts, manifestDesc ocispec.Descriptor, reference string) (*ocispec.Manifest, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference: %w", err)
	}

	remoteRepo, err := getRepositoryClient(ctx, secureOpts, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository client: %w", err)
	}

	fetcher, err := registryutil.NewBlobFetcher(ctx, reference, remoteRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to create blob fetcher: %w", err)
	}

	manifest, err := fetcher.FetchManifest(ctx, manifestDesc)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}

	return manifest, nil
}

func runSign(command *cobra.Command, cmdOpts *signOpts) error {
	// set log level
	ctx := cmdOpts.LoggingFlagOpts.InitializeLogger(command.Context())

	// initialize
	var signer sign.Signer
	var err error
	signer, err = sign.GetSigner(ctx, &cmdOpts.SignerFlagOpts)
	if err != nil {
		return err
	}
	sigRepo, err := getRepository(ctx, cmdOpts.inputType, cmdOpts.reference, &cmdOpts.SecureFlagOpts, cmdOpts.forceReferrersTag)
	if err != nil {
		return err
	}
	signOpts, err := prepareSigningOpts(ctx, cmdOpts)
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
	var artifactManifestDesc, sigManifestDesc ocispec.Descriptor
	if cmdOpts.dmVerity {
		// dm-verity flow: sign layers with PKCS#7, push signatures, sign manifest with JWS/COSE
		manifest, err := fetchImageManifest(ctx, &cmdOpts.SecureFlagOpts, manifestDesc, cmdOpts.reference)
		if err != nil {
			return fmt.Errorf("failed to fetch manifest: %w", err)
		}

		ref, err := registry.ParseReference(cmdOpts.reference)
		if err != nil {
			return fmt.Errorf("failed to parse reference: %w", err)
		}
		remoteRepo, err := getRepositoryClient(ctx, &cmdOpts.SecureFlagOpts, ref)
		if err != nil {
			return fmt.Errorf("failed to get repository client: %w", err)
		}
		blobFetcher, err := registryutil.NewBlobFetcher(ctx, cmdOpts.reference, remoteRepo)
		if err != nil {
			return fmt.Errorf("failed to create blob fetcher: %w", err)
		}

		primitiveSigner, err := sign.GetPrimitiveSigner(ctx, &cmdOpts.SignerFlagOpts)
		if err != nil {
			return fmt.Errorf("failed to get primitive signer for dm-verity: %w", err)
		}

		layerSignatures, err := dmverity.SignImageLayers(ctx, primitiveSigner, blobFetcher, *manifest)
		if err != nil {
			return fmt.Errorf("failed to sign layers with dm-verity: %w", err)
		}

		sigManifest, err := dmverity.CreateSignatureManifest(layerSignatures, manifestDesc)
		if err != nil {
			return fmt.Errorf("failed to create signature manifest: %w", err)
		}

		manifestJSON, err := json.MarshalIndent(sigManifest, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal signature manifest: %w", err)
		}

		manifestDigest := digest.FromBytes(manifestJSON)

		logger := log.GetLogger(ctx)
		separator := strings.Repeat("=", 100)
		logger.Debugf("\n%s\n", separator)
		logger.Debugf("Digest: %s\n", manifestDigest)
		logger.Debugf("Size: %d bytes\n", len(manifestJSON))
		logger.Debugf("%s\n", separator)
		logger.Debugf("%s\n", string(manifestJSON))
		logger.Debugf("%s\n\n", separator)

		layerSigManifestDesc, err := pushDmVerityManifest(ctx, remoteRepo, sigManifest, layerSignatures)
		if err != nil {
			return fmt.Errorf("failed to push dm-verity manifest: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Pushed dm-verity layer signatures: %s\n", layerSigManifestDesc.Digest)
	}

	artifactManifestDesc, sigManifestDesc, err = notation.SignOCI(ctx, signer, sigRepo, signOpts)
	if err != nil {
		var referrerError *remote.ReferrersError
		if !errors.As(err, &referrerError) || !referrerError.IsReferrersIndexDelete() {
			return err
		}
		// show warning for referrers index deletion failed
		fmt.Fprintln(os.Stderr, "Warning: Removal of outdated referrers index from remote registry failed. Garbage collection may be required.")
	}

	repositoryRef, _, _ := strings.Cut(resolvedRef, "@")
	fmt.Printf("Successfully signed %s@%s\n", repositoryRef, artifactManifestDesc.Digest.String())
	fmt.Printf("Pushed the signature to %s@%s\n", repositoryRef, sigManifestDesc.Digest.String())
	return nil
}

func prepareSigningOpts(ctx context.Context, opts *signOpts) (notation.SignOptions, error) {
	logger := log.GetLogger(ctx)

	mediaType, err := envelope.GetEnvelopeMediaType(opts.SignerFlagOpts.SignatureFormat)
	if err != nil {
		return notation.SignOptions{}, err
	}
	pluginConfig, err := flag.ParseFlagMap(opts.pluginConfig, flag.PflagPluginConfig.Name)
	if err != nil {
		return notation.SignOptions{}, err
	}
	userMetadata, err := flag.ParseFlagMap(opts.userMetadata, flag.PflagUserMetadata.Name)
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
	if opts.tsaServerURL != "" {
		// timestamping
		logger.Infof("Configured to timestamp with TSA %q", opts.tsaServerURL)
		signOpts.Timestamper, err = tspclient.NewHTTPTimestamper(httputil.NewClient(ctx, &http.Client{Timeout: timestampingTimeout}), opts.tsaServerURL)
		if err != nil {
			return notation.SignOptions{}, fmt.Errorf("cannot get http timestamper for timestamping: %w", err)
		}
		signOpts.TSARootCAs, err = nx509.NewRootCertPool(opts.tsaRootCertificatePath)
		if err != nil {
			return notation.SignOptions{}, err
		}
		tsaRevocationValidator, err := clirev.NewRevocationValidator(ctx, purpose.Timestamping)
		if err != nil {
			return notation.SignOptions{}, fmt.Errorf("failed to create timestamping revocation validator: %w", err)
		}
		signOpts.TSARevocationValidator = tsaRevocationValidator
	}
	return signOpts, nil
}

// pushDmVerityManifest pushes the dm-verity signature manifest and layer signatures to the registry, returns descriptor of the pushed manifest.
// It pushes the layer signatures as blobs, then pushes the manifest referencing those blobs. It verifies that all blobs are present in the registry before pushing the manifest to avoid a broken state where the manifest is pushed with missing blobs.
func pushDmVerityManifest(ctx context.Context, repo registry.Repository, sigManifest *dmverity.SignatureManifest, layerSignatures []dmverity.SignatureEnvelope) (ocispec.Descriptor, error) {
	manifestJSON, err := json.MarshalIndent(sigManifest, "", "  ")
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("failed to marshal signature manifest: %w", err)
	}

	manifestDigest := digest.FromBytes(manifestJSON)
	manifestDesc := ocispec.Descriptor{
		MediaType: sigManifest.MediaType,
		Digest:    manifestDigest,
		Size:      int64(len(manifestJSON)),
	}

	// Push the empty config blob (OCI spec requires config even for artifacts)
	emptyConfig := []byte("{}")
	emptyConfigDesc := sigManifest.Config

	// Try to push; ignore if already exists
	err = repo.Blobs().Push(ctx, emptyConfigDesc, bytes.NewReader(emptyConfig))
	if err != nil {
		_, statErr := repo.Blobs().Resolve(ctx, emptyConfigDesc.Digest.String())
		if statErr != nil {
			return ocispec.Descriptor{}, fmt.Errorf("failed to push empty config blob: push: %w, stat: %v", err, statErr)
		}
	}

	for i, sig := range layerSignatures {
		desc := sigManifest.Layers[i]
		err := repo.Blobs().Push(ctx, desc, bytes.NewReader(sig.Signature))
		if err != nil {
			_, statErr := repo.Blobs().Resolve(ctx, desc.Digest.String())
			if statErr != nil {
				return ocispec.Descriptor{}, fmt.Errorf("failed to push signature blob for layer %s: push: %w, stat: %v", sig.LayerDigest, err, statErr)
			}
		}
	}

	// Verify all blobs are present before pushing manifest
	if _, err := repo.Blobs().Resolve(ctx, emptyConfigDesc.Digest.String()); err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("empty config blob missing before manifest push: %w", err)
	}
	for i := range layerSignatures {
		desc := sigManifest.Layers[i]
		if _, err := repo.Blobs().Resolve(ctx, desc.Digest.String()); err != nil {
			return ocispec.Descriptor{}, fmt.Errorf("signature blob %s missing before manifest push: %w", desc.Digest, err)
		}
	}

	err = repo.Manifests().Push(ctx, manifestDesc, bytes.NewReader(manifestJSON))
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("failed to push signature manifest: %w", err)
	}

	return manifestDesc, nil
}
