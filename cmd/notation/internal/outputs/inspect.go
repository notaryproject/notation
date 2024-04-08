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

package outputs

import (
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/notaryproject/notation/internal/tree"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"os"
	"strconv"
	"strings"
	"time"
)

type InspectOutput struct {
	MediaType    string `json:"mediaType"`
	Signatures   []SignatureOutput
	outputFormat string
}

type SignatureOutput struct {
	MediaType             string              `json:"mediaType"`
	Digest                string              `json:"digest"`
	SignatureAlgorithm    string              `json:"signatureAlgorithm"`
	SignedAttributes      map[string]string   `json:"signedAttributes"`
	UserDefinedAttributes map[string]string   `json:"userDefinedAttributes"`
	UnsignedAttributes    map[string]string   `json:"unsignedAttributes"`
	Certificates          []certificateOutput `json:"certificates"`
	SignedArtifact        ocispec.Descriptor  `json:"signedArtifact"`
}

type certificateOutput struct {
	SHA256Fingerprint string `json:"SHA256Fingerprint"`
	IssuedTo          string `json:"issuedTo"`
	IssuedBy          string `json:"issuedBy"`
	Expiry            string `json:"expiry"`
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

func getUnsignedAttributes(envContent *signature.EnvelopeContent) map[string]string {
	unsignedAttributes := map[string]string{}

	if envContent.SignerInfo.UnsignedAttributes.TimestampSignature != nil {
		unsignedAttributes["timestampSignature"] = b64.StdEncoding.EncodeToString(envContent.SignerInfo.UnsignedAttributes.TimestampSignature)
	}

	if envContent.SignerInfo.UnsignedAttributes.SigningAgent != "" {
		unsignedAttributes["signingAgent"] = envContent.SignerInfo.UnsignedAttributes.SigningAgent
	}

	return unsignedAttributes
}

func getCertificates(outputFormat string, envContent *signature.EnvelopeContent) []certificateOutput {
	certificates := []certificateOutput{}

	for _, cert := range envContent.SignerInfo.CertificateChain {
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

func addMapToTree(node *tree.Node, m map[string]string) {
	if len(m) > 0 {
		for k, v := range m {
			node.AddPair(k, v)
		}
	} else {
		node.Add("(empty)")
	}
}

func formatTimestamp(outputFormat string, t time.Time) string {
	switch outputFormat {
	case cmd.OutputJSON:
		return t.Format(time.RFC3339)
	default:
		return t.Format(time.ANSIC)
	}
}

func Signatures(mediaType string, digest string, output InspectOutput, sigFile []byte) (error, []SignatureOutput) {
	sigEnvelope, err := signature.ParseEnvelope(mediaType, sigFile)
	skippedSignatures := false
	if err != nil {
		logSkippedSignature(digest, err)
		skippedSignatures = true
		return nil, nil
	}

	envelopeContent, err := sigEnvelope.Content()
	if err != nil {
		logSkippedSignature(digest, err)
		skippedSignatures = true
		return nil, nil
	}

	signedArtifactDesc, err := envelope.DescriptorFromSignaturePayload(&envelopeContent.Payload)
	if err != nil {
		logSkippedSignature(digest, err)
		skippedSignatures = true
		return nil, nil
	}

	signatureAlgorithm, err := proto.EncodeSigningAlgorithm(envelopeContent.SignerInfo.SignatureAlgorithm)
	if err != nil {
		logSkippedSignature(digest, err)
		skippedSignatures = true
		return nil, nil
	}

	sig := SignatureOutput{
		MediaType:             mediaType,
		Digest:                digest,
		SignatureAlgorithm:    string(signatureAlgorithm),
		SignedAttributes:      getSignedAttributes(InspectOutput{}.outputFormat, envelopeContent),
		UserDefinedAttributes: signedArtifactDesc.Annotations,
		UnsignedAttributes:    getUnsignedAttributes(envelopeContent),
		Certificates:          getCertificates(InspectOutput{}.outputFormat, envelopeContent),
		SignedArtifact:        *signedArtifactDesc,
	}

	// clearing annotations from the SignedArtifact field since they're already
	// displayed as UserDefinedAttributes
	sig.SignedArtifact.Annotations = nil

	output.Signatures = append(output.Signatures, sig)

	if skippedSignatures {
		return errors.New("at least one signature was skipped and not displayed"), nil
	}
	return nil, output.Signatures
}
func PrintOutput(outputFormat string, ref string, output InspectOutput) error {
	if outputFormat == cmd.OutputJSON {
		return ioutil.PrintObjectAsJSON(output)
	}

	if len(output.Signatures) == 0 {
		fmt.Printf("%s has no associated signature\n", ref)
		return nil
	}

	fmt.Println("Inspecting all signatures for signed artifact")
	root := tree.New(ref)

	for _, signature := range output.Signatures {

		if !(signature.MediaType == "jws" || signature.MediaType == "cose") {
			subroot := root.Add(registry.ArtifactTypeNotation)
			subroot.Add(signature.Digest)
		}
		root.AddPair("media type", signature.MediaType)
		root.AddPair("signature algorithm", signature.SignatureAlgorithm)

		signedAttributesNode := root.Add("signed attributes")
		addMapToTree(signedAttributesNode, signature.SignedAttributes)

		userDefinedAttributesNode := root.Add("user defined attributes")
		addMapToTree(userDefinedAttributesNode, signature.UserDefinedAttributes)

		unsignedAttributesNode := root.Add("unsigned attributes")
		addMapToTree(unsignedAttributesNode, signature.UnsignedAttributes)

		certListNode := root.Add("certificates")
		for _, cert := range signature.Certificates {
			certNode := certListNode.AddPair("SHA256 fingerprint", cert.SHA256Fingerprint)
			certNode.AddPair("issued to", cert.IssuedTo)
			certNode.AddPair("issued by", cert.IssuedBy)
			certNode.AddPair("expiry", cert.Expiry)
		}

		artifactNode := root.Add("signed artifact")
		artifactNode.AddPair("media type", signature.SignedArtifact.MediaType)
		artifactNode.AddPair("digest", signature.SignedArtifact.Digest.String())
		artifactNode.AddPair("size", strconv.FormatInt(signature.SignedArtifact.Size, 10))
	}

	root.Print()

	return nil
}

func logSkippedSignature(digest string, err error) {
	fmt.Fprintf(os.Stderr, "Warning: Skipping signature %s because of error: %v\n", digest, err)
}
