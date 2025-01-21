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

package tree

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/cmd/notation/internal/output"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/internal/tree"
	"github.com/notaryproject/tspclient-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// InspectHandler is a handler for inspecting metadata information and rendering
// it in a tree format.
type InspectHandler struct {
	printer *output.Printer

	// rootReferenceNode is the root node with the artifact reference as the
	// value.
	rootReferenceNode *tree.Node
	// notationSignaturesNode is the node for all signatures associated with the
	// artifact.
	notationSignaturesNode *tree.Node
}

// NewInspectHandler creates a new InspectHandler.
func NewInspectHandler(printer *output.Printer) *InspectHandler {
	return &InspectHandler{
		printer: printer,
	}
}

// OnReferenceResolved sets the artifact reference and media type for the
// handler.
//
// mediaType is a no-op for this handler.
func (h *InspectHandler) OnReferenceResolved(reference, _ string) {
	h.rootReferenceNode = tree.New(reference)
	h.notationSignaturesNode = h.rootReferenceNode.Add(registry.ArtifactTypeNotation)
}

// InspectSignature inspects a signature to get it ready to be rendered.
func (h *InspectHandler) InspectSignature(manifestDesc ocispec.Descriptor, envelope signature.Envelope) error {
	return addSignature(h.notationSignaturesNode, manifestDesc.Digest.String(), envelope)
}

// Render renders the metadata information when an operation is complete.
func (h *InspectHandler) Render() error {
	if len(h.notationSignaturesNode.Children) == 0 {
		return h.printer.Printf("%s has no associated signature\n", h.rootReferenceNode.Value)
	}
	h.printer.Println("Inspecting all signatures for signed artifact")
	return h.rootReferenceNode.Print(h.printer)
}

func addSignature(node *tree.Node, digest string, sigEnvelope signature.Envelope) error {
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

	// create signature node
	sigNode := node.Add(digest)
	sigNode.AddPair("signature algorithm", string(signatureAlgorithm))

	addSignedAttributes(sigNode, envelopeContent)
	addUserDefinedAttributes(sigNode, signedArtifactDesc.Annotations)
	addUnsignedAttributes(sigNode, envelopeContent)
	addCertificates(sigNode, envelopeContent.SignerInfo.CertificateChain)
	addSignedArtifact(sigNode, signedArtifactDesc)
	return nil
}

func addSignedAttributes(node *tree.Node, envelopeContent *signature.EnvelopeContent) {
	signedAttributesNode := node.Add("signed attributes")
	signedAttributesNode.AddPair("signing scheme", string(envelopeContent.SignerInfo.SignedAttributes.SigningScheme))
	signedAttributesNode.AddPair("signing time", formatTime(envelopeContent.SignerInfo.SignedAttributes.SigningTime))
	if expiry := envelopeContent.SignerInfo.SignedAttributes.Expiry; !expiry.IsZero() {
		signedAttributesNode.AddPair("expiry", formatTime(expiry))
	}
	for _, attribute := range envelopeContent.SignerInfo.SignedAttributes.ExtendedAttributes {
		signedAttributesNode.AddPair(fmt.Sprint(attribute.Key), fmt.Sprint(attribute.Value))
	}
}

func addUserDefinedAttributes(node *tree.Node, annotations map[string]string) {
	userDefinedAttributesNode := node.Add("user defined attributes")
	if len(annotations) == 0 {
		userDefinedAttributesNode.Add("(empty)")
		return
	}
	for _, k := range orderedKeys(annotations) {
		v := annotations[k]
		userDefinedAttributesNode.AddPair(k, v)
	}
}

func addUnsignedAttributes(node *tree.Node, envelopeContent *signature.EnvelopeContent) {
	unsignedAttributesNode := node.Add("unsigned attributes")
	if signingAgent := envelopeContent.SignerInfo.UnsignedAttributes.SigningAgent; signingAgent != "" {
		unsignedAttributesNode.AddPair("signing agent", signingAgent)
	}
	if timestamp := envelopeContent.SignerInfo.UnsignedAttributes.TimestampSignature; timestamp != nil {
		addTimestamp(unsignedAttributesNode, envelopeContent.SignerInfo)
	}
}

func addSignedArtifact(node *tree.Node, signedArtifactDesc ocispec.Descriptor) {
	artifactNode := node.Add("signed artifact")
	artifactNode.AddPair("media type", signedArtifactDesc.MediaType)
	artifactNode.AddPair("digest", signedArtifactDesc.Digest.String())
	artifactNode.AddPair("size", strconv.FormatInt(signedArtifactDesc.Size, 10))
}

func addTimestamp(node *tree.Node, signerInfo signature.SignerInfo) {
	timestampNode := node.Add("timestamp signature")
	signedToken, err := tspclient.ParseSignedToken(signerInfo.UnsignedAttributes.TimestampSignature)
	if err != nil {
		timestampNode.AddPair("error", fmt.Sprintf("failed to parse timestamp countersignature: %s", err))
		return
	}
	info, err := signedToken.Info()
	if err != nil {
		timestampNode.AddPair("error", fmt.Sprintf("failed to parse timestamp countersignature: %s", err))
		return
	}
	timestamp, err := info.Validate(signerInfo.Signature)
	if err != nil {
		timestampNode.AddPair("error", fmt.Sprintf("failed to parse timestamp countersignature: %s", err))
		return
	}
	timestampNode.AddPair("timestamp", timestamp.Format(time.ANSIC))
	addCertificates(timestampNode, signedToken.Certificates)
}

func addCertificates(node *tree.Node, certChain []*x509.Certificate) {
	certListNode := node.Add("certificates")
	for _, cert := range certChain {
		hash := sha256.Sum256(cert.Raw)
		certNode := certListNode.AddPair("SHA256 fingerprint", strings.ToLower(hex.EncodeToString(hash[:])))
		certNode.AddPair("issued to", cert.Subject.String())
		certNode.AddPair("issued by", cert.Issuer.String())
		certNode.AddPair("expiry", formatTime(cert.NotAfter))
	}
}

func orderedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func formatTime(t time.Time) string {
	return t.Format(time.ANSIC)
}
