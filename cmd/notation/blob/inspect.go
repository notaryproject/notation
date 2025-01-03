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
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-core-go/signature/cose"
	"github.com/notaryproject/notation-core-go/signature/jws"
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/notaryproject/notation/internal/tree"
	"github.com/notaryproject/tspclient-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

type inspectOpts struct {
	sigPath      string
	outputFormat string
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

func inspectCommand() *cobra.Command {
	opts := &inspectOpts{}
	command := &cobra.Command{
		Use:   "inspect [flags] <signature_path>",
		Short: "Inspect a signature associated with a blob",
		Long: `Inspect a signature associated with a blob.

Example - Inspect a signature:
  notation inspect blob.cose.sig
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing signature path: use `notation blob inspect --help` to see what parameters are required")
			}
			opts.sigPath = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInspect(opts)
		},
	}

	cmd.SetPflagOutput(command.Flags(), &opts.outputFormat, cmd.PflagOutputUsage)
	return command
}

func runInspect(opts *inspectOpts) error {
	if opts.outputFormat != cmd.OutputJSON && opts.outputFormat != cmd.OutputPlaintext {
		return fmt.Errorf("unrecognized output format %s", opts.outputFormat)
	}

	envelopeMediaType, err := parseEnvelopeMediaType(filepath.Base(opts.sigPath))
	if err != nil {
		return err
	}

	sigBlob, err := os.ReadFile(opts.sigPath)
	if err != nil {
		return err
	}

	sigEnvelope, err := signature.ParseEnvelope(envelopeMediaType, sigBlob)
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

	sig := signatureOutput{
		MediaType:             envelopeMediaType,
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

	return printOutput(opts.sigPath, sig, opts.outputFormat)
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

func printOutput(sigPath string, signature signatureOutput, outputFormat string) error {
	if outputFormat == cmd.OutputJSON {
		return ioutil.PrintObjectAsJSON(signature)
	}

	sigNode := tree.New(sigPath)
	sigNode.AddPair("signature algorithm", signature.SignatureAlgorithm)
	sigNode.AddPair("signature envelope type", signature.MediaType)

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

	sigNode.Print()
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

func parseEnvelopeMediaType(filename string) (string, error) {
	parts := strings.Split(filename, ".")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid signature filename: %s", filename)
	}
	mediaType := strings.ToLower(parts[len(parts)-2])
	switch mediaType {
	case "jws":
		return jws.MediaTypeEnvelope, nil
	case "cose":
		return cose.MediaTypeEnvelope, nil
	default:
		return "", fmt.Errorf("unsupported signature format: %s", mediaType)
	}
}
