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
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/notation-go/registry"
	cmderr "github.com/notaryproject/notation/cmd/notation/internal/errors"
	"github.com/notaryproject/notation/cmd/notation/internal/experimental"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/notaryproject/notation/internal/tree"
	"github.com/notaryproject/tspclient-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

type inspectOpts struct {
	cmd.LoggingFlagOpts
	SecureFlagOpts
	reference         string
	outputFormat      string
	allowReferrersAPI bool
	maxSignatures     int
}

type inspectOutput struct {
	MediaType  string `json:"mediaType"`
	Signatures []signatureOutput
}

type signatureOutput struct {
	MediaType             string              `json:"mediaType"`
	Digest                string              `json:"digest"`
	SignatureAlgorithm    string              `json:"signatureAlgorithm"`
	SignedAttributes      map[string]string   `json:"signedAttributes"`
	UserDefinedAttributes map[string]string   `json:"userDefinedAttributes"`
	UnsignedAttributes    map[string]any      `json:"unsignedAttributes"`
	Certificates          []certificateOutput `json:"certificates"`
	SignedArtifact        ocispec.Descriptor  `json:"signedArtifact"`
}

type certificateOutput struct {
	SHA256Fingerprint string `json:"SHA256Fingerprint"`
	IssuedTo          string `json:"issuedTo"`
	IssuedBy          string `json:"issuedBy"`
	Expiry            string `json:"expiry"`
}

type timestampOutput struct {
	Timestamp    string              `json:"timestamp,omitempty"`
	Certificates []certificateOutput `json:"certificates,omitempty"`
	Error        string              `json:"error,omitempty"`
}

func inspectCommand(opts *inspectOpts) *cobra.Command {
	if opts == nil {
		opts = &inspectOpts{}
	}
	longMessage := `Inspect all signatures associated with the signed artifact.

Example - Inspect signatures on an OCI artifact identified by a digest:
  notation inspect <registry>/<repository>@<digest>

Example - Inspect signatures on an OCI artifact identified by a tag  (Notation will resolve tag to digest):
  notation inspect <registry>/<repository>:<tag>

Example - Inspect signatures on an OCI artifact identified by a digest and output as json:
  notation inspect --output json <registry>/<repository>@<digest>
`
	command := &cobra.Command{
		Use:   "inspect [reference]",
		Short: "Inspect all signatures associated with the signed artifact",
		Long:  longMessage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing reference to the artifact: use `notation inspect --help` to see what parameters are required")
			}
			opts.reference = args[0]
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return experimental.CheckFlagsAndWarn(cmd, "allow-referrers-api")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.maxSignatures <= 0 {
				return fmt.Errorf("max-signatures value %d must be a positive number", opts.maxSignatures)
			}
			if cmd.Flags().Changed("allow-referrers-api") {
				fmt.Fprintln(os.Stderr, "Warning: flag '--allow-referrers-api' is deprecated and will be removed in future versions.")
			}
			return runInspect(cmd, opts)
		},
	}

	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	cmd.SetPflagOutput(command.Flags(), &opts.outputFormat, cmd.PflagOutputUsage)
	command.Flags().IntVar(&opts.maxSignatures, "max-signatures", 100, "maximum number of signatures to evaluate or examine")
	cmd.SetPflagReferrersAPI(command.Flags(), &opts.allowReferrersAPI, fmt.Sprintf(cmd.PflagReferrersUsageFormat, "inspect"))
	return command
}

func runInspect(command *cobra.Command, opts *inspectOpts) error {
	// set log level
	ctx := opts.LoggingFlagOpts.InitializeLogger(command.Context())

	if opts.outputFormat != cmd.OutputJSON && opts.outputFormat != cmd.OutputPlaintext {
		return fmt.Errorf("unrecognized output format %s", opts.outputFormat)
	}

	// initialize
	reference := opts.reference
	// always use the Referrers API, if not supported, automatically fallback to
	// the referrers tag schema
	sigRepo, err := getRemoteRepository(ctx, &opts.SecureFlagOpts, reference, false)
	if err != nil {
		return err
	}
	manifestDesc, resolvedRef, err := resolveReferenceWithWarning(ctx, inputTypeRegistry, reference, sigRepo, "inspect")
	if err != nil {
		return err
	}
	output := inspectOutput{MediaType: manifestDesc.MediaType, Signatures: []signatureOutput{}}
	skippedSignatures := false
	err = listSignatures(ctx, sigRepo, manifestDesc, opts.maxSignatures, func(sigManifestDesc ocispec.Descriptor) error {
		sigBlob, sigDesc, err := sigRepo.FetchSignatureBlob(ctx, sigManifestDesc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: unable to fetch signature %s due to error: %v\n", sigManifestDesc.Digest.String(), err)
			skippedSignatures = true
			return nil
		}

		sigEnvelope, err := signature.ParseEnvelope(sigDesc.MediaType, sigBlob)
		if err != nil {
			logSkippedSignature(sigManifestDesc, err)
			skippedSignatures = true
			return nil
		}

		envelopeContent, err := sigEnvelope.Content()
		if err != nil {
			logSkippedSignature(sigManifestDesc, err)
			skippedSignatures = true
			return nil
		}

		signedArtifactDesc, err := envelope.DescriptorFromSignaturePayload(&envelopeContent.Payload)
		if err != nil {
			logSkippedSignature(sigManifestDesc, err)
			skippedSignatures = true
			return nil
		}

		signatureAlgorithm, err := proto.EncodeSigningAlgorithm(envelopeContent.SignerInfo.SignatureAlgorithm)
		if err != nil {
			logSkippedSignature(sigManifestDesc, err)
			skippedSignatures = true
			return nil
		}

		sig := signatureOutput{
			MediaType:             sigDesc.MediaType,
			Digest:                sigManifestDesc.Digest.String(),
			SignatureAlgorithm:    string(signatureAlgorithm),
			SignedAttributes:      getSignedAttributes(opts.outputFormat, envelopeContent),
			UserDefinedAttributes: signedArtifactDesc.Annotations,
			UnsignedAttributes:    getUnsignedAttributes(opts.outputFormat, envelopeContent),
			Certificates:          getCertificates(opts.outputFormat, envelopeContent.SignerInfo.CertificateChain),
			SignedArtifact:        *signedArtifactDesc,
		}

		// clearing annotations from the SignedArtifact field since they're already
		// displayed as UserDefinedAttributes
		sig.SignedArtifact.Annotations = nil

		output.Signatures = append(output.Signatures, sig)

		return nil
	})
	var errorExceedMaxSignatures cmderr.ErrorExceedMaxSignatures
	if err != nil && !errors.As(err, &errorExceedMaxSignatures) {
		return err
	}

	if err := printOutput(opts.outputFormat, resolvedRef, output); err != nil {
		return err
	}

	if errorExceedMaxSignatures.MaxSignatures > 0 {
		fmt.Println("Warning:", errorExceedMaxSignatures)
	}

	if skippedSignatures {
		return errors.New("at least one signature was skipped and not displayed")
	}

	return nil
}

func logSkippedSignature(sigDesc ocispec.Descriptor, err error) {
	fmt.Fprintf(os.Stderr, "Warning: Skipping signature %s because of error: %v\n", sigDesc.Digest.String(), err)
}

func getSignedAttributes(outputFormat string, envContent *signature.EnvelopeContent) map[string]string {
	signedAttributes := map[string]string{
		"signingScheme": string(envContent.SignerInfo.SignedAttributes.SigningScheme),
		"signingTime":   formatTimestamp(outputFormat, envContent.SignerInfo.SignedAttributes.SigningTime),
	}
	expiry := envContent.SignerInfo.SignedAttributes.Expiry
	if !expiry.IsZero() {
		signedAttributes["expiry"] = formatTimestamp(outputFormat, expiry)
	}

	for _, attribute := range envContent.SignerInfo.SignedAttributes.ExtendedAttributes {
		signedAttributes[fmt.Sprint(attribute.Key)] = fmt.Sprint(attribute.Value)
	}

	return signedAttributes
}

func getUnsignedAttributes(outputFormat string, envContent *signature.EnvelopeContent) map[string]any {
	unsignedAttributes := make(map[string]any)

	if envContent.SignerInfo.UnsignedAttributes.TimestampSignature != nil {
		unsignedAttributes["timestampSignature"] = parseTimestamp(outputFormat, envContent.SignerInfo)
	}

	if envContent.SignerInfo.UnsignedAttributes.SigningAgent != "" {
		unsignedAttributes["signingAgent"] = envContent.SignerInfo.UnsignedAttributes.SigningAgent
	}

	return unsignedAttributes
}

func formatTimestamp(outputFormat string, t time.Time) string {
	switch outputFormat {
	case cmd.OutputJSON:
		return t.Format(time.RFC3339)
	default:
		return t.Format(time.ANSIC)
	}
}

func getCertificates(outputFormat string, certChain []*x509.Certificate) []certificateOutput {
	certificates := []certificateOutput{}

	for _, cert := range certChain {
		h := sha256.Sum256(cert.Raw)
		fingerprint := strings.ToLower(hex.EncodeToString(h[:]))

		certificate := certificateOutput{
			SHA256Fingerprint: fingerprint,
			IssuedTo:          cert.Subject.String(),
			IssuedBy:          cert.Issuer.String(),
			Expiry:            formatTimestamp(outputFormat, cert.NotAfter),
		}

		certificates = append(certificates, certificate)
	}

	return certificates
}

func printOutput(outputFormat string, ref string, output inspectOutput) error {
	if outputFormat == cmd.OutputJSON {
		return ioutil.PrintObjectAsJSON(output)
	}

	if len(output.Signatures) == 0 {
		fmt.Printf("%s has no associated signature\n", ref)
		return nil
	}

	fmt.Println("Inspecting all signatures for signed artifact")
	root := tree.New(ref)
	cncfSigNode := root.Add(registry.ArtifactTypeNotation)

	for _, signature := range output.Signatures {
		sigNode := cncfSigNode.Add(signature.Digest)
		sigNode.AddPair("media type", signature.MediaType)
		sigNode.AddPair("signature algorithm", signature.SignatureAlgorithm)

		signedAttributesNode := sigNode.Add("signed attributes")
		addMapToTree(signedAttributesNode, signature.SignedAttributes)

		userDefinedAttributesNode := sigNode.Add("user defined attributes")
		addMapToTree(userDefinedAttributesNode, signature.UserDefinedAttributes)

		unsignedAttributesNode := sigNode.Add("unsigned attributes")
		for k, v := range signature.UnsignedAttributes {
			switch value := v.(type) {
			case string:
				unsignedAttributesNode.AddPair(k, value)
			case timestampOutput:
				timestampNode := unsignedAttributesNode.Add("timestamp signature")
				if value.Error != "" {
					timestampNode.AddPair("error", value.Error)
					break
				}
				timestampNode.AddPair("timestamp", value.Timestamp)
				addCertificatesToTree(timestampNode, "certificates", value.Certificates)
			}
		}

		addCertificatesToTree(sigNode, "certificates", signature.Certificates)

		artifactNode := sigNode.Add("signed artifact")
		artifactNode.AddPair("media type", signature.SignedArtifact.MediaType)
		artifactNode.AddPair("digest", signature.SignedArtifact.Digest.String())
		artifactNode.AddPair("size", strconv.FormatInt(signature.SignedArtifact.Size, 10))
	}

	root.Print()
	return nil
}

func addMapToTree(node *tree.Node, m map[string]string) {
	if len(m) > 0 {
		for k, v := range m {
			node.AddPair(k, v)
		}
	} else {
		node.Add("(empty)")
	}
}

func addCertificatesToTree(node *tree.Node, name string, certs []certificateOutput) {
	certListNode := node.Add(name)
	for _, cert := range certs {
		certNode := certListNode.AddPair("SHA256 fingerprint", cert.SHA256Fingerprint)
		certNode.AddPair("issued to", cert.IssuedTo)
		certNode.AddPair("issued by", cert.IssuedBy)
		certNode.AddPair("expiry", cert.Expiry)
	}
}

func parseTimestamp(outputFormat string, signerInfo signature.SignerInfo) timestampOutput {
	signedToken, err := tspclient.ParseSignedToken(signerInfo.UnsignedAttributes.TimestampSignature)
	if err != nil {
		return timestampOutput{
			Error: fmt.Sprintf("failed to parse timestamp countersignature: %s", err.Error()),
		}
	}
	info, err := signedToken.Info()
	if err != nil {
		return timestampOutput{
			Error: fmt.Sprintf("failed to parse timestamp countersignature: %s", err.Error()),
		}
	}
	timestamp, err := info.Validate(signerInfo.Signature)
	if err != nil {
		return timestampOutput{
			Error: fmt.Sprintf("failed to parse timestamp countersignature: %s", err.Error()),
		}
	}
	certificates := getCertificates(outputFormat, signedToken.Certificates)
	var formatTimestamp string
	switch outputFormat {
	case cmd.OutputJSON:
		formatTimestamp = timestamp.Format(time.RFC3339)
	default:
		formatTimestamp = timestamp.Format(time.ANSIC)
	}
	return timestampOutput{
		Timestamp:    formatTimestamp,
		Certificates: certificates,
	}
}
