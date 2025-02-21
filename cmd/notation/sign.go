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
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/notaryproject/notation-core-go/revocation/purpose"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/log"
	"github.com/notaryproject/notation/cmd/notation/internal/experimental"
	"github.com/notaryproject/notation/cmd/notation/internal/option"
	"github.com/notaryproject/notation/cmd/notation/internal/signer"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/internal/httputil"
	clirev "github.com/notaryproject/notation/internal/revocation"
	nx509 "github.com/notaryproject/notation/internal/x509"
	"github.com/notaryproject/tspclient-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

const referrersTagSchemaDeleteError = "failed to delete dangling referrers index"

// timestampingTimeout is the timeout when requesting timestamp countersignature
// from a TSA
const timestampingTimeout = 15 * time.Second

type signOpts struct {
	option.Logging
	option.Signer
	option.UserMetadata
	SecureFlagOpts
	expiry                 time.Duration
	reference              string
	allowReferrersAPI      bool
	forceReferrersTag      bool
	ociLayout              bool
	inputType              inputType
	tsaServerURL           string
	tsaRootCertificatePath string
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
`
	experimentalExamples := `
Example - [Experimental] Sign an OCI artifact referenced in an OCI layout
  notation sign --oci-layout "<oci_layout_path>@<digest>"

Example - [Experimental] Sign an OCI artifact identified by a tag and referenced in an OCI layout
  notation sign --oci-layout "<oci_layout_path>:<tag>"
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
			return experimental.CheckFlagsAndWarn(cmd, "allow-referrers-api", "oci-layout")
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

			// allow-referrers-api flag is set
			if cmd.Flags().Changed("allow-referrers-api") {
				if opts.allowReferrersAPI {
					fmt.Fprintln(os.Stderr, "Warning: flag '--allow-referrers-api' is deprecated and will be removed in future versions, use '--force-referrers-tag=false' instead.")
					opts.forceReferrersTag = false
				} else {
					fmt.Fprintln(os.Stderr, "Warning: flag '--allow-referrers-api' is deprecated and will be removed in future versions.")
				}
			}
			return runSign(cmd, opts)
		},
	}
	fs := command.Flags()
	opts.Logging.ApplyFlags(fs)
	opts.Signer.ApplyFlags(command)
	opts.SecureFlagOpts.ApplyFlags(fs)
	cmd.SetPflagExpiry(command.Flags(), &opts.expiry)
	opts.UserMetadata.ApplyFlags(fs)
	cmd.SetPflagReferrersAPI(fs, &opts.allowReferrersAPI, fmt.Sprintf(cmd.PflagReferrersUsageFormat, "sign"))
	command.Flags().StringVar(&opts.tsaServerURL, "timestamp-url", "", "RFC 3161 Timestamping Authority (TSA) server URL")
	command.Flags().StringVar(&opts.tsaRootCertificatePath, "timestamp-root-cert", "", "filepath of timestamp authority root certificate")
	cmd.SetPflagReferrersTag(fs, &opts.forceReferrersTag, "force to store signatures using the referrers tag schema")
	command.Flags().BoolVar(&opts.ociLayout, "oci-layout", false, "[Experimental] sign the artifact stored as OCI image layout")
	command.MarkFlagsMutuallyExclusive("oci-layout", "force-referrers-tag", "allow-referrers-api")
	command.MarkFlagsRequiredTogether("timestamp-url", "timestamp-root-cert")
	experimental.HideFlags(command, experimentalExamples, []string{"oci-layout"})
	return command
}

func runSign(command *cobra.Command, opts *signOpts) error {
	// set log level
	ctx := opts.Logging.InitializeLogger(command.Context())

	// initialize
	signer, err := signer.GetSigner(ctx, &opts.Signer)
	if err != nil {
		return err
	}
	sigRepo, err := getRepository(ctx, opts.inputType, opts.reference, &opts.SecureFlagOpts, opts.forceReferrersTag)
	if err != nil {
		return err
	}
	signOpts, err := prepareSigningOpts(ctx, opts)
	if err != nil {
		return err
	}
	manifestDesc, resolvedRef, err := resolveReference(ctx, opts.inputType, opts.reference, sigRepo, func(ref string, manifestDesc ocispec.Descriptor) {
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
		if errors.As(err, &errorPushSignatureFailed) && strings.Contains(err.Error(), referrersTagSchemaDeleteError) {
			fmt.Fprintln(os.Stderr, "Warning: Removal of outdated referrers index from remote registry failed. Garbage collection may be required.")
			// write out
			fmt.Println("Successfully signed", resolvedRef)
			return nil
		}
		return err
	}
	fmt.Println("Successfully signed", resolvedRef)
	return nil
}

func prepareSigningOpts(ctx context.Context, opts *signOpts) (notation.SignOptions, error) {
	logger := log.GetLogger(ctx)

	mediaType, err := envelope.GetEnvelopeMediaType(opts.Signer.SignatureFormat)
	if err != nil {
		return notation.SignOptions{}, err
	}
	pluginConfig, err := opts.Signer.PluginConfigMap()
	if err != nil {
		return notation.SignOptions{}, err
	}
	userMetadata, err := opts.UserMetadataMap()
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
