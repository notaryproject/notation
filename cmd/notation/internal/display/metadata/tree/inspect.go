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

type InspectHandler struct {
	root        *tree.Node
	cncfSigNode *tree.Node
	printer     *output.Printer
}

func NewInspectHandler(printer *output.Printer) *InspectHandler {
	return &InspectHandler{
		printer: printer,
	}
}

// SetReference sets the artifact reference for the handler.
func (h *InspectHandler) SetReference(reference string) {
	if h.root == nil {
		h.root = tree.New(reference)
		h.cncfSigNode = h.root.Add(registry.ArtifactTypeNotation)
	}
}

// SetMediaType sets the media type for the handler. It is a no-op for this
// handler.
func (h *InspectHandler) SetMediaType(_ string) {}

func (h *InspectHandler) InspectSignature(digest string, envelopeMediaType string, sigEnvelope signature.Envelope) error {
	if h.root == nil || h.cncfSigNode == nil {
		return fmt.Errorf("artifact reference is not set")
	}

	return addSignature(h.cncfSigNode, digest, envelopeMediaType, sigEnvelope)
}

func (h *InspectHandler) Print() error {
	if h.root == nil {
		return fmt.Errorf("artifact reference is not set")
	}
	return h.root.Print(h.printer)
}

func addSignature(node *tree.Node, digest string, envelopeMediaType string, sigEnvelope signature.Envelope) error {
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
	sigNode.AddPair("signature envelope type", envelopeMediaType)

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

func addSignedArtifact(node *tree.Node, signedArtifactDesc *ocispec.Descriptor) {
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
