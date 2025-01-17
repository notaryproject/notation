package model

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/tspclient-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Signature is the signature envelope for printing in human readable format.
type Signature struct {
	MediaType             string             `json:"mediaType"`
	Digest                string             `json:"digest,omitempty"`
	SignatureAlgorithm    string             `json:"signatureAlgorithm"`
	SignedAttributes      map[string]string  `json:"signedAttributes"`
	UserDefinedAttributes map[string]string  `json:"userDefinedAttributes"`
	UnsignedAttributes    map[string]any     `json:"unsignedAttributes"`
	Certificates          []Certificate      `json:"certificates"`
	SignedArtifact        ocispec.Descriptor `json:"signedArtifact"`
}

// Certificate is the certificate information for printing in human readable
// format.
type Certificate struct {
	SHA256Fingerprint string `json:"SHA256Fingerprint"`
	IssuedTo          string `json:"issuedTo"`
	IssuedBy          string `json:"issuedBy"`
	Expiry            string `json:"expiry"`
}

// Timestamp is the timestamp information for printing in human readable.
type Timestamp struct {
	Timestamp    string        `json:"timestamp,omitempty"`
	Certificates []Certificate `json:"certificates,omitempty"`
	Error        string        `json:"error,omitempty"`
}

// formatter is the function to format the value to string.
type formatter func(any) string

// NewSignature parses the signature blob and returns a Signature object.
func NewSignature(digest string, envelopeMediaType string, sigEnvelope signature.Envelope, formatter formatter) (*Signature, error) {
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
		SignedAttributes:      getSignedAttributes(envelopeContent, formatter),
		UserDefinedAttributes: signedArtifactDesc.Annotations,
		UnsignedAttributes:    getUnsignedAttributes(envelopeContent, formatter),
		Certificates:          getCertificates(envelopeContent.SignerInfo.CertificateChain, formatter),
		SignedArtifact:        *signedArtifactDesc,
	}

	// clearing annotations from the SignedArtifact field since they're already
	// displayed as UserDefinedAttributes
	sig.SignedArtifact.Annotations = nil

	return sig, nil
}

func getSignedAttributes(envContent *signature.EnvelopeContent, formatter formatter) map[string]string {
	signedAttributes := map[string]string{
		"signingScheme": string(envContent.SignerInfo.SignedAttributes.SigningScheme),
		"signingTime":   formatter(envContent.SignerInfo.SignedAttributes.SigningTime),
	}
	if expiry := envContent.SignerInfo.SignedAttributes.Expiry; !expiry.IsZero() {
		signedAttributes["expiry"] = formatter(expiry)
	}

	for _, attribute := range envContent.SignerInfo.SignedAttributes.ExtendedAttributes {
		signedAttributes[fmt.Sprint(attribute.Key)] = fmt.Sprint(attribute.Value)
	}
	return signedAttributes
}

func getUnsignedAttributes(envContent *signature.EnvelopeContent, formatter formatter) map[string]any {
	unsignedAttributes := make(map[string]any)

	if envContent.SignerInfo.UnsignedAttributes.TimestampSignature != nil {
		unsignedAttributes["timestampSignature"] = parseTimestamp(envContent.SignerInfo, formatter)
	}

	if envContent.SignerInfo.UnsignedAttributes.SigningAgent != "" {
		unsignedAttributes["signingAgent"] = envContent.SignerInfo.UnsignedAttributes.SigningAgent
	}
	return unsignedAttributes
}

func getCertificates(certChain []*x509.Certificate, formatter formatter) []Certificate {
	certificates := []Certificate{}

	for _, cert := range certChain {
		hash := sha256.Sum256(cert.Raw)
		certificates = append(certificates, Certificate{
			SHA256Fingerprint: strings.ToLower(hex.EncodeToString(hash[:])),
			IssuedTo:          cert.Subject.String(),
			IssuedBy:          cert.Issuer.String(),
			Expiry:            formatter(cert.NotAfter),
		})
	}
	return certificates
}

func parseTimestamp(signerInfo signature.SignerInfo, formatter formatter) Timestamp {
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
		Timestamp:    formatter(*timestamp),
		Certificates: getCertificates(signedToken.Certificates, formatter),
	}
}
