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

package blob

import (
	"errors"
	"fmt"
	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/notation/cmd/notation/internal/outputs"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

type blobInspectOpts struct {
	cmd.LoggingFlagOpts
	cmd.SignerFlagOpts
	signaturePath string
	outputFormat  string
	mediaType     string
}

func inspectCommand(opts *blobInspectOpts) *cobra.Command {
	if opts == nil {
		opts = &blobInspectOpts{}
	}
	longMessage := `Inspect signature associated with the signed blob.

Example - Inspect BLOB signature:
  notation blob inspect <signature_path>

Example - Inspect BLOB signature and output as JSON:
  notation blob inspect --output json <signature_path>
`

	command := &cobra.Command{
		Use:   "blob inspect [signaturePath]",
		Short: "Inspect signature associated with the signed BLOB",
		Long:  longMessage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing signature path to the artifact: use `notation blob inspect --help` to see what parameters are required")
			}
			opts.signaturePath = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBlobInspect(opts)
		},
	}

	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	cmd.SetPflagOutput(command.Flags(), &opts.outputFormat, cmd.PflagOutputUsage)
	return command
}

func runBlobInspect(opts *blobInspectOpts) error {
	if opts.outputFormat != cmd.OutputJSON && opts.outputFormat != cmd.OutputPlaintext {
		return fmt.Errorf("unrecognized output format %s", opts.outputFormat)
	}

	// initialize
	mediaType, err := envelope.GetEnvelopeMediaType(filepath.Ext(opts.signaturePath))
	if !(mediaType == "jws.MediaTypeEnvelope" || mediaType == "cose.MediaTypeEnvelope") {
		return err
	}
	contents, err := os.ReadFile(opts.signaturePath)
	if err != nil {
		return err
	}
	output := outputs.InspectOutput{MediaType: mediaType, Signatures: []outputs.SignatureOutput{}}
	sigEnvelope, err := signature.ParseEnvelope(mediaType, contents)
	if err != nil {
		return err
	}
	envelopeContent, err := sigEnvelope.Content()
	if err != nil {
		return err
	}
	signedArtifactDesc, err := envelope.DescriptorFromSignaturePayload(&envelopeContent.Payload)
	if err != nil {
		return err
	}
	signatureAlgorithm, err := proto.EncodeSigningAlgorithm(envelopeContent.SignerInfo.SignatureAlgorithm)
	if err != nil {
		return err
	}

	sig := outputs.SignatureOutput{
		MediaType:             mediaType,
		SignatureAlgorithm:    string(signatureAlgorithm),
		SignedAttributes:      outputs.GetSignedAttributes(opts.outputFormat, envelopeContent),
		UserDefinedAttributes: signedArtifactDesc.Annotations,
		UnsignedAttributes:    outputs.GetUnsignedAttributes(envelopeContent),
		Certificates:          outputs.GetCertificates(opts.outputFormat, envelopeContent),
		SignedArtifact:        *signedArtifactDesc,
	}

	sig.SignedArtifact.Annotations = nil

	output.Signatures = append(output.Signatures, sig)
	if err := outputs.PrintOutput(opts.outputFormat, opts.signaturePath, output); err != nil {
		return err
	}
	return nil
}
