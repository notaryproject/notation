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

package json

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	coresignature "github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go/plugin/proto"
	envelopeutil "github.com/notaryproject/notation/v2/internal/envelope"
	"github.com/notaryproject/tspclient-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// signature is the signature envelope for printing in JSON format.
type signature struct {
	Digest                string             `json:"digest,omitempty"`
	SignatureAlgorithm    string             `json:"signatureAlgorithm"`
	SignatureEnvelopeType string             `json:"signatureEnvelopeType"`
	SignedAttributes      map[string]any     `json:"signedAttributes"`
	UserDefinedAttributes map[string]string  `json:"userDefinedAttributes"`
	UnsignedAttributes    map[string]any     `json:"unsignedAttributes"`
	Certificates          []*certificate     `json:"certificates"`
	SignedArtifact        ocispec.Descriptor `json:"signedArtifact"`
}

// certificate is the certificate information for printing in JSON format.
type certificate struct {
	SHA256Fingerprint string    `json:"SHA256Fingerprint"`
	IssuedTo          string    `json:"issuedTo"`
	IssuedBy          string    `json:"issuedBy"`
	Expiry            time.Time `json:"expiry"`
}

// timestamp is the timestamp information for printing in JSON format.
type timestamp struct {
	Timestamp    string         `json:"timestamp,omitempty"`
	Certificates []*certificate `json:"certificates,omitempty"`
	Error        string         `json:"error,omitempty"`
}

// newSignature parses the signature blob and returns a Signature object.
func newSignature(digest, signatureMediaType string, envelope coresignature.Envelope) (*signature, error) {
	envelopeContent, err := envelope.Content()
	if err != nil {
		return nil, err
	}

	signedArtifactDesc, err := envelopeutil.DescriptorFromSignaturePayload(&envelopeContent.Payload)
	if err != nil {
		return nil, err
	}

	signatureAlgorithm, err := proto.EncodeSigningAlgorithm(envelopeContent.SignerInfo.SignatureAlgorithm)
	if err != nil {
		return nil, err
	}
	sig := &signature{
		Digest:                digest,
		SignatureEnvelopeType: signatureMediaType,
		SignatureAlgorithm:    string(signatureAlgorithm),
		SignedAttributes:      getSignedAttributes(envelopeContent),
		UserDefinedAttributes: signedArtifactDesc.Annotations,
		UnsignedAttributes:    getUnsignedAttributes(envelopeContent),
		Certificates:          getCertificates(envelopeContent.SignerInfo.CertificateChain),
		SignedArtifact:        signedArtifactDesc,
	}

	// clearing annotations from the SignedArtifact field since they're already
	// displayed as UserDefinedAttributes
	sig.SignedArtifact.Annotations = nil

	return sig, nil
}

func getSignedAttributes(envelopeContent *coresignature.EnvelopeContent) map[string]any {
	signedAttributes := map[string]any{
		"contentType":   envelopeContent.Payload.ContentType,
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

func getUnsignedAttributes(envelopeContent *coresignature.EnvelopeContent) map[string]any {
	unsignedAttributes := make(map[string]any)
	if envelopeContent.SignerInfo.UnsignedAttributes.SigningAgent != "" {
		unsignedAttributes["signingAgent"] = envelopeContent.SignerInfo.UnsignedAttributes.SigningAgent
	}
	if envelopeContent.SignerInfo.UnsignedAttributes.TimestampSignature != nil {
		unsignedAttributes["timestampSignature"] = parseTimestamp(envelopeContent.SignerInfo)
	}
	return unsignedAttributes
}

func getCertificates(certChain []*x509.Certificate) []*certificate {
	var certificates []*certificate
	for _, cert := range certChain {
		hash := sha256.Sum256(cert.Raw)
		certificates = append(certificates, &certificate{
			SHA256Fingerprint: strings.ToLower(hex.EncodeToString(hash[:])),
			IssuedTo:          cert.Subject.String(),
			IssuedBy:          cert.Issuer.String(),
			Expiry:            cert.NotAfter,
		})
	}
	return certificates
}

func parseTimestamp(signerInfo coresignature.SignerInfo) *timestamp {
	signedToken, err := tspclient.ParseSignedToken(signerInfo.UnsignedAttributes.TimestampSignature)
	if err != nil {
		return &timestamp{
			Error: fmt.Sprintf("failed to parse timestamp countersignature: %s", err),
		}
	}
	info, err := signedToken.Info()
	if err != nil {
		return &timestamp{
			Error: fmt.Sprintf("failed to parse timestamp countersignature: %s", err),
		}
	}
	ts, err := info.Validate(signerInfo.Signature)
	if err != nil {
		return &timestamp{
			Error: fmt.Sprintf("failed to parse timestamp countersignature: %s", err),
		}
	}
	return &timestamp{
		Timestamp:    ts.Format(time.RFC3339Nano),
		Certificates: getCertificates(signedToken.Certificates),
	}
}
