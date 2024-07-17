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
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	corex509 "github.com/notaryproject/notation-core-go/x509"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/cmd/notation/internal/experimental"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/internal/httputil"
	"github.com/notaryproject/tspclient-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

const referrersTagSchemaDeleteError = "failed to delete dangling referrers index"

// timestampingTimeout is the timeout when requesting timestamp countersignature
// from a TSA
const timestampingTimeout = 15 * time.Second

type signOpts struct {
	cmd.LoggingFlagOpts
	cmd.SignerFlagOpts
	SecureFlagOpts
	expiry                 time.Duration
	pluginConfig           []string
	userMetadata           []string
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
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SignerFlagOpts.ApplyFlagsToCommand(command)
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	cmd.SetPflagExpiry(command.Flags(), &opts.expiry)
	cmd.SetPflagPluginConfig(command.Flags(), &opts.pluginConfig)
	cmd.SetPflagUserMetadata(command.Flags(), &opts.userMetadata, cmd.PflagUserMetadataSignUsage)
	cmd.SetPflagReferrersAPI(command.Flags(), &opts.allowReferrersAPI, fmt.Sprintf(cmd.PflagReferrersUsageFormat, "sign"))
	command.Flags().StringVar(&opts.tsaServerURL, "timestamp-url", "", "RFC 3161 Timestamping Authority (TSA) server URL")
	command.Flags().StringVar(&opts.tsaRootCertificatePath, "timestamp-root-cert", "", "filepath of timestamp authority root certificate")
	cmd.SetPflagReferrersTag(command.Flags(), &opts.forceReferrersTag, "force to store signatures using the referrers tag schema")
	command.Flags().BoolVar(&opts.ociLayout, "oci-layout", false, "[Experimental] sign the artifact stored as OCI image layout")
	command.MarkFlagsMutuallyExclusive("oci-layout", "force-referrers-tag", "allow-referrers-api")
	command.MarkFlagsRequiredTogether("timestamp-url", "timestamp-root-cert")
	experimental.HideFlags(command, experimentalExamples, []string{"oci-layout"})
	return command
}

func runSign(command *cobra.Command, cmdOpts *signOpts) error {
	// set log level
	ctx := cmdOpts.LoggingFlagOpts.InitializeLogger(command.Context())

	// initialize
	signer, err := cmd.GetSigner(ctx, &cmdOpts.SignerFlagOpts)
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
	if opts.tsaServerURL != "" {
		// timestamping
		fmt.Printf("Configured to timestamp with TSA %q\n", opts.tsaServerURL)
		signOpts.Timestamper, err = tspclient.NewHTTPTimestamper(httputil.NewClient(ctx, &http.Client{Timeout: timestampingTimeout}), opts.tsaServerURL)
		if err != nil {
			return notation.SignOptions{}, fmt.Errorf("cannot get http timestamper for timestamping: %w", err)
		}

		rootCerts, err := corex509.ReadCertificateFile(opts.tsaRootCertificatePath)
		if err != nil {
			return notation.SignOptions{}, err
		}
		if len(rootCerts) == 0 {
			return notation.SignOptions{}, fmt.Errorf("cannot find any tsa root certificate from %q. Expecting x509 certificate in PEM or DER format from the file", opts.tsaRootCertificatePath)
		}
		rootCAs := x509.NewCertPool()
		rootCAs.AddCert(rootCerts[0])
		signOpts.TSARootCAs = rootCAs
	}
	return signOpts, nil
}
