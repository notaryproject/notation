package json

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/notation/cmd/notation/internal/output"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/tspclient-go"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type inspectOutput struct {
	MediaType  string `json:"mediaType"`
	Signatures []Signature
}

// Signature is the signature envelope for printing in human readable format.
type Signature struct {
	MediaType             string             `json:"mediaType"`
	Digest                string             `json:"digest,omitempty"`
	SignatureAlgorithm    string             `json:"signatureAlgorithm"`
	SignedAttributes      map[string]any     `json:"signedAttributes"`
	UserDefinedAttributes map[string]string  `json:"userDefinedAttributes"`
	UnsignedAttributes    map[string]any     `json:"unsignedAttributes"`
	Certificates          []Certificate      `json:"certificates"`
	SignedArtifact        ocispec.Descriptor `json:"signedArtifact"`
}

// Certificate is the certificate information for printing in human readable
// format.
type Certificate struct {
	SHA256Fingerprint string    `json:"SHA256Fingerprint"`
	IssuedTo          string    `json:"issuedTo"`
	IssuedBy          string    `json:"issuedBy"`
	Expiry            time.Time `json:"expiry"`
}

// Timestamp is the timestamp information for printing in human readable.
type Timestamp struct {
	Timestamp    string        `json:"timestamp,omitempty"`
	Certificates []Certificate `json:"certificates,omitempty"`
	Error        string        `json:"error,omitempty"`
}

type InspectHandler struct {
	output inspectOutput

	// printer is the printer for output.
	printer *output.Printer
}

// NewInspectHandler creates a new InspectHandler.
func NewInspectHandler(printer *output.Printer) *InspectHandler {
	return &InspectHandler{
		printer: printer,
	}
}

// SetReference sets the artifact reference for the handler. It is a no-op for this
// handler.
func (h *InspectHandler) SetReference(_ string) {}

// SetMediaType sets the media type for the handler.
func (h *InspectHandler) SetMediaType(mediaType string) {
	if h.output.MediaType == "" {
		h.output.MediaType = mediaType
	}
}

// InspectSignature inspects a signature to get it ready to be rendered.
func (h *InspectHandler) InspectSignature(digest string, envelopeMediaType string, sigEnvelope signature.Envelope) error {
	sig, err := newSignature(digest, envelopeMediaType, sigEnvelope)
	if err != nil {
		return err
	}
	h.output.Signatures = append(h.output.Signatures, *sig)
	return nil
}

func (h *InspectHandler) Print() error {
	return output.PrintPrettyJSON(h.printer, h.output)
}

// newSignature parses the signature blob and returns a Signature object.
func newSignature(digest string, envelopeMediaType string, sigEnvelope signature.Envelope) (*Signature, error) {
	envelopeContent, err := sigEnvelope.Content()
	if err != nil {
		return nil, err
	}

	signedArtifactDesc, err := envelope.DescriptorFromSignaturePayload(&envelopeContent.Payload)
	if err != nil {
		return nil, err
	}

	signatureAlgorithm, err := proto.EncodeSigningAlgorithm(envelopeContent.SignerInfo.SignatureAlgorithm)
	if err != nil {
		return nil, err
	}
	sig := &Signature{
		MediaType:             envelopeMediaType,
		Digest:                digest,
		SignatureAlgorithm:    string(signatureAlgorithm),
		SignedAttributes:      getSignedAttributes(envelopeContent),
		UserDefinedAttributes: signedArtifactDesc.Annotations,
		UnsignedAttributes:    getUnsignedAttributes(envelopeContent),
		Certificates:          getCertificates(envelopeContent.SignerInfo.CertificateChain),
		SignedArtifact:        *signedArtifactDesc,
	}

	// clearing annotations from the SignedArtifact field since they're already
	// displayed as UserDefinedAttributes
	sig.SignedArtifact.Annotations = nil

	return sig, nil
}

func getSignedAttributes(envelopeContent *signature.EnvelopeContent) map[string]any {
	signedAttributes := map[string]any{
		"signingScheme": string(envelopeContent.SignerInfo.SignedAttributes.SigningScheme),
		"signingTime":   envelopeContent.SignerInfo.SignedAttributes.SigningTime,
	}
	if expiry := envelopeContent.SignerInfo.SignedAttributes.Expiry; !expiry.IsZero() {
		signedAttributes["expiry"] = expiry
	}
	for _, attribute := range envelopeContent.SignerInfo.SignedAttributes.ExtendedAttributes {
		signedAttributes[fmt.Sprint(attribute.Key)] = fmt.Sprint(attribute.Value)
	}
	return signedAttributes
}

func getUnsignedAttributes(envelopeContent *signature.EnvelopeContent) map[string]any {
	unsignedAttributes := make(map[string]any)
	if envelopeContent.SignerInfo.UnsignedAttributes.SigningAgent != "" {
		unsignedAttributes["signingAgent"] = envelopeContent.SignerInfo.UnsignedAttributes.SigningAgent
	}
	if envelopeContent.SignerInfo.UnsignedAttributes.TimestampSignature != nil {
		unsignedAttributes["timestampSignature"] = parseTimestamp(envelopeContent.SignerInfo)
	}
	return unsignedAttributes
}

func getCertificates(certChain []*x509.Certificate) []Certificate {
	var certificates []Certificate
	for _, cert := range certChain {
		hash := sha256.Sum256(cert.Raw)
		certificates = append(certificates, Certificate{
			SHA256Fingerprint: strings.ToLower(hex.EncodeToString(hash[:])),
			IssuedTo:          cert.Subject.String(),
			IssuedBy:          cert.Issuer.String(),
			Expiry:            cert.NotAfter,
		})
	}
	return certificates
}

func parseTimestamp(signerInfo signature.SignerInfo) Timestamp {
	signedToken, err := tspclient.ParseSignedToken(signerInfo.UnsignedAttributes.TimestampSignature)
	if err != nil {
		return Timestamp{
			Error: fmt.Sprintf("failed to parse timestamp countersignature: %s", err),
		}
	}
	info, err := signedToken.Info()
	if err != nil {
		return Timestamp{
			Error: fmt.Sprintf("failed to parse timestamp countersignature: %s", err),
		}
	}
	timestamp, err := info.Validate(signerInfo.Signature)
	if err != nil {
		return Timestamp{
			Error: fmt.Sprintf("failed to parse timestamp countersignature: %s", err),
		}
	}
	return Timestamp{
		Timestamp:    timestamp.Format(time.RFC3339Nano),
		Certificates: getCertificates(signedToken.Certificates),
	}
}
