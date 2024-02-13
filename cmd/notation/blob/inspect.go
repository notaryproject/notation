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
	"github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/cmd/notation/internal/sharedutils"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/notaryproject/notation/internal/tree"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"strconv"
)

type blobInspectOpts struct {
	cmd.LoggingFlagOpts
	cmd.SignerFlagOpts
	desc          ocispec.Descriptor
	sigRepo       registry.Repository
	signaturePath string
	outputFormat  string
	MediaType     string
}

func inspectCommand(opts *blobInspectOpts) *cobra.Command {
	if opts == nil {
		opts = &blobInspectOpts{}
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
				return errors.New("missing signature path to the artifact: use `notation blob inspect --help` to see what parameters are required")
			}
			opts.signaturePath = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBlobInspect(cmd, opts)
		},
	}

	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	cmd.SetPflagOutput(command.Flags(), &opts.outputFormat, cmd.PflagOutputUsage)
	return command
}

func runBlobInspect(command *cobra.Command, opts *blobInspectOpts) error {
	// set log level
	ctx := opts.LoggingFlagOpts.InitializeLogger(command.Context())

	if opts.outputFormat != cmd.OutputJSON && opts.outputFormat != cmd.OutputPlaintext {
		return fmt.Errorf("unrecognized output format %s", opts.outputFormat)
	}

	output := sharedutils.InspectOutput{MediaType: opts.MediaType, Signatures: []sharedutils.SignatureOutput{}}
	//Added sigRepo and desc as placeholders. Once notation-go changes are merged revisit and update FetchSignatureBlob
	sigBlob, _, _ := opts.sigRepo.FetchSignatureBlob(ctx, opts.desc)
	sigEnvelope, _ := signature.ParseEnvelope(opts.MediaType, sigBlob)
	envelopeContent, _ := sigEnvelope.Content()
	signedArtifactDesc, _ := envelope.DescriptorFromSignaturePayload(&envelopeContent.Payload)
	signatureAlgorithm, _ := proto.EncodeSigningAlgorithm(envelopeContent.SignerInfo.SignatureAlgorithm)

	sig := sharedutils.SignatureOutput{
		SignatureAlgorithm:    string(signatureAlgorithm),
		SignedAttributes:      sharedutils.GetSignedAttributes(opts.outputFormat, envelopeContent),
		UserDefinedAttributes: signedArtifactDesc.Annotations,
		UnsignedAttributes:    sharedutils.GetUnsignedAttributes(envelopeContent),
		Certificates:          sharedutils.GetCertificates(opts.outputFormat, envelopeContent),
		SignedArtifact:        *signedArtifactDesc,
	}

	// clearing annotations from the SignedArtifact field since they're already
	// displayed as UserDefinedAttributes
	sig.SignedArtifact.Annotations = nil

	output.Signatures = append(output.Signatures, sig)
	blobPath := "placeholder"
	if err := printOutput(opts.outputFormat, blobPath, output); err != nil {
		return err
	}
	return nil
}

func printOutput(outputFormat string, blobPath string, output sharedutils.InspectOutput) error {
	if outputFormat == cmd.OutputJSON {
		return ioutil.PrintObjectAsJSON(output)
	}

	if len(output.Signatures) == 0 {
		fmt.Printf("%s has no associated signature\n", blobPath)
		return nil
	}

	fmt.Println("Inspecting all signatures for signed artifact")
	root := tree.New(blobPath)
	cncfSigNode := root.Add(registry.ArtifactTypeNotation)

	for _, signature := range output.Signatures {
		sigNode := cncfSigNode.Add(signature.Digest)
		sigNode.AddPair("media type", signature.MediaType)
		sigNode.AddPair("signature algorithm", signature.SignatureAlgorithm)

		signedAttributesNode := sigNode.Add("signed attributes")
		sharedutils.AddMapToTree(signedAttributesNode, signature.SignedAttributes)

		userDefinedAttributesNode := sigNode.Add("user defined attributes")
		sharedutils.AddMapToTree(userDefinedAttributesNode, signature.UserDefinedAttributes)

		unsignedAttributesNode := sigNode.Add("unsigned attributes")
		sharedutils.AddMapToTree(unsignedAttributesNode, signature.UnsignedAttributes)

		certListNode := sigNode.Add("certificates")
		for _, cert := range signature.Certificates {
			certNode := certListNode.AddPair("SHA256 fingerprint", cert.SHA256Fingerprint)
			certNode.AddPair("issued to", cert.IssuedTo)
			certNode.AddPair("issued by", cert.IssuedBy)
			certNode.AddPair("expiry", cert.Expiry)
		}

		artifactNode := sigNode.Add("signed artifact")
		artifactNode.AddPair("media type", signature.SignedArtifact.MediaType)
		artifactNode.AddPair("digest", signature.SignedArtifact.Digest.String())
		artifactNode.AddPair("size", strconv.FormatInt(signature.SignedArtifact.Size, 10))
	}

	root.Print()
	return nil
}
