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

//Verify this
import (
	"context"
	"errors"
	"fmt"
	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

type blobOpts struct {
	cmd.LoggingFlagOpts
	cmd.SignerFlagOpts
	SecureFlagOpts
	expiry        time.Duration
	desc          ocispec.Descriptor
	sigRepo       registry.Repository
	pluginConfig  []string
	userMetadata  []string
	blobPath      string
	signaturePath string
	outputFormat  string
}

func blobSignCommand(opts *blobOpts) *cobra.Command {
	if opts == nil {
		opts = &blobOpts{}
	}
	longMessage := `Sign BLOB artifacts

Note: a signing key must be specified. This can be done temporarily by specifying a key ID, or a new key can be configured using the command "notation key add"

Example - Sign a BLOB artifact using the default signing key, with the default JWS envelope, and use BLOB image manifest to store the signature:
  notation blob sign <blob_path>

Example - Sign a BLOB artifact by generating the signature in a particular directory: 
 notation blob sign --signature-directory <directory_path> <blob_path>

Example - Sign a BLOB artifact and skip user confirmations when overwriting existing signature:
  notation blob sign --force <blob_path> 

Example - Sign a BLOB artifact using the default signing key, with the COSE envelope:
  notation blob sign --signature-format cose <blob_path>

Example - Sign a BLOB artifact with a specified plugin and signing key stored in KMS: 
  notation blob sign --plugin <plugin_name> --id <remote_key_id> <blob_path>

Example - Sign a BLOB artifact and add a user metadata to payload: 
  notation blob sign --user-metadata <metadata> <blob_path>

Example - Sign a BLOB artifact using a specified media type: 
  notation blob sign --media-type <media type> <blob_path>

Example - Sign a BLOB artifact using a specified key: 
  notation blob sign --key <key_name> <blob_path>

Example - Sign a BLOB artifact and specify the signature expiry duration, for example 24 hours: 
  notation blob sign --expiry 24h <blob_path>
`

	command := &cobra.Command{
		Use:   "blob sign [flags] <blobPath>",
		Short: "Sign BLOB artifacts",
		Long:  longMessage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing blob_path")
			}
			opts.blobPath = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBlobSign(cmd, opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SignerFlagOpts.ApplyFlagsToCommand(command)
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	cmd.SetPflagExpiry(command.Flags(), &opts.expiry)
	cmd.SetPflagPluginConfig(command.Flags(), &opts.pluginConfig)
	cmd.SetPflagUserMetadata(command.Flags(), &opts.userMetadata, cmd.PflagUserMetadataSignUsage)
	//PlaceHolder for MediaType and Signature-directory
	return command
}

func runBlobSign(command *cobra.Command, cmdOpts *blobOpts) error {
	// set log level
	ctx := cmdOpts.LoggingFlagOpts.InitializeLogger(command.Context())

	// initialize
	signer, err := cmd.GetSigner(ctx, &cmdOpts.SignerFlagOpts)
	if err != nil {
		return err
	}
	blobOpts, err := prepareBlobSigningOpts(ctx, cmdOpts)
	if err != nil {
		return err
	}

	// core process
	err = notation.BlobSign(ctx, signer, blobOpts) //PlaceHolder
	if err != nil {
		var errorPushSignatureFailed notation.ErrorPushSignatureFailed
		if errors.As(err, &errorPushSignatureFailed) && strings.Contains(err.Error(), referrersTagSchemaDeleteError) {
			fmt.Fprintln(os.Stderr, "Warning: Removal of outdated referrers index from remote registry failed. Garbage collection may be required.")
			// write out
			fmt.Println("Successfully signed")
			return nil
		}
		return err
	}
	fmt.Println("Successfully signed")
	return nil
}

func prepareBlobSigningOpts(ctx context.Context, opts *blobOpts) (notation.SignOptions, error) {
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
	blobOpts := notation.SignOptions{
		SignerSignOptions: notation.SignerSignOptions{
			SignatureMediaType: mediaType,
			ExpiryDuration:     opts.expiry,
			PluginConfig:       pluginConfig,
		},
		UserMetadata: userMetadata,
	}
	return blobOpts, nil
}

func blobInspectCommand(opts *blobOpts) *cobra.Command {
	if opts == nil {
		opts = &blobOpts{}
	}
	longMessage := `Inspect all signatures associated with the signed artifact.

Example - Inspect signatures on an BLOB artifact:
  notation blob inspect <signature_path>

Example - Inspect signatures on an BLOB artifact output as json:
  notation blob inspect --output json <signature_path>
`

	command := &cobra.Command{
		Use:   "blob inspect [signaturePath]",
		Short: "Inspect all signatures associated with the signed artifact",
		Long:  longMessage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing signature_path")
			}
			opts.signaturePath = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBlobInspect(cmd, opts)
		},
	}

	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	cmd.SetPflagOutput(command.Flags(), &opts.outputFormat, cmd.PflagOutputUsage)
	return command
}

func runBlobInspect(command *cobra.Command, opts *blobOpts) error {
	// set log level
	ctx := opts.LoggingFlagOpts.InitializeLogger(command.Context())

	if opts.outputFormat != cmd.OutputJSON && opts.outputFormat != cmd.OutputPlaintext {
		return fmt.Errorf("unrecognized output format %s", opts.outputFormat)
	}

	output := inspectOutput{MediaType: opts.desc.MediaType, Signatures: []signatureOutput{}}

	sigBlob, _, _ := opts.sigRepo.FetchSignatureBlob(ctx, opts.desc)
	sigEnvelope, _ := signature.ParseEnvelope(opts.desc.MediaType, sigBlob)
	envelopeContent, _ := sigEnvelope.Content()
	signedArtifactDesc, _ := envelope.DescriptorFromSignaturePayload(&envelopeContent.Payload)
	signatureAlgorithm, _ := proto.EncodeSigningAlgorithm(envelopeContent.SignerInfo.SignatureAlgorithm)

	sig := signatureOutput{
		MediaType:             opts.desc.MediaType,
		Digest:                opts.desc.Digest.String(),
		SignatureAlgorithm:    string(signatureAlgorithm),
		SignedAttributes:      getSignedAttributes(opts.outputFormat, envelopeContent),
		UserDefinedAttributes: signedArtifactDesc.Annotations,
		UnsignedAttributes:    getUnsignedAttributes(envelopeContent),
		Certificates:          getCertificates(opts.outputFormat, envelopeContent),
		SignedArtifact:        *signedArtifactDesc,
	}

	// clearing annotations from the SignedArtifact field since they're already
	// displayed as UserDefinedAttributes
	sig.SignedArtifact.Annotations = nil

	output.Signatures = append(output.Signatures, sig)

	return nil
}
