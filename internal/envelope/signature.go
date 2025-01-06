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

package envelope

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/notaryproject/notation-plugin-framework-go/plugin"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/notaryproject/notation/internal/tree"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go/plugin/proto"
	"github.com/notaryproject/tspclient-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// SignatureInfo is the signature envelope with human readable fields.
type SignatureInfo struct {
	MediaType             string                    `json:"mediaType"`
	Digest                string                    `json:"digest,omitempty"`
	SignatureAlgorithm    plugin.SignatureAlgorithm `json:"signatureAlgorithm"`
	SignedAttributes      map[string]any            `json:"signedAttributes"`
	UserDefinedAttributes map[string]string         `json:"userDefinedAttributes"`
	UnsignedAttributes    map[string]any            `json:"unsignedAttributes"`
	Certificates          []CertificateInfo         `json:"certificates"`
	SignedArtifact        ocispec.Descriptor        `json:"signedArtifact"`
}

type CertificateInfo struct {
	SHA256Fingerprint string      `json:"SHA256Fingerprint"`
	IssuedTo          string      `json:"issuedTo"`
	IssuedBy          string      `json:"issuedBy"`
	Expiry            ioutil.Time `json:"expiry"`
}

type TimestampInfo struct {
	Timestamp    ioutil.Timestamp  `json:"timestamp,omitempty"`
	Certificates []CertificateInfo `json:"certificates,omitempty"`
	Error        string            `json:"error,omitempty"`
}

func Parse(sig []byte, envelopeMediaType string) (*SignatureInfo, error) {
	sigEnvelope, err := signature.ParseEnvelope(envelopeMediaType, sig)
	if err != nil {
		return nil, err
	}

	envelopeContent, err := sigEnvelope.Content()
	if err != nil {
		return nil, err
	}

	signedArtifactDesc, err := DescriptorFromSignaturePayload(&envelopeContent.Payload)
	if err != nil {
		return nil, err
	}

	signatureAlgorithm, err := proto.EncodeSigningAlgorithm(envelopeContent.SignerInfo.SignatureAlgorithm)
	if err != nil {
		return nil, err
	}

	return &SignatureInfo{
		MediaType:             envelopeMediaType,
		SignatureAlgorithm:    signatureAlgorithm,
		SignedAttributes:      getSignedAttributes(envelopeContent),
		UserDefinedAttributes: signedArtifactDesc.Annotations,
		UnsignedAttributes:    getUnsignedAttributes(envelopeContent),
		Certificates:          getCertificates(envelopeContent.SignerInfo.CertificateChain),
		SignedArtifact:        *signedArtifactDesc,
	}, nil
}

func getSignedAttributes(envContent *signature.EnvelopeContent) map[string]any {
	signedAttributes := map[string]any{
		"signingScheme": envContent.SignerInfo.SignedAttributes.SigningScheme,
		"signingTime":   ioutil.Time(envContent.SignerInfo.SignedAttributes.SigningTime),
	}
	expiry := envContent.SignerInfo.SignedAttributes.Expiry
	if !expiry.IsZero() {
		signedAttributes["expiry"] = ioutil.Time(expiry)
	}

	for _, attribute := range envContent.SignerInfo.SignedAttributes.ExtendedAttributes {
		signedAttributes[fmt.Sprint(attribute.Key)] = fmt.Sprint(attribute.Value)
	}

	return signedAttributes
}

func getUnsignedAttributes(envContent *signature.EnvelopeContent) map[string]any {
	unsignedAttributes := make(map[string]any)

	if envContent.SignerInfo.UnsignedAttributes.TimestampSignature != nil {
		unsignedAttributes["timestampSignature"] = parseTimestamp(envContent.SignerInfo)
	}

	if envContent.SignerInfo.UnsignedAttributes.SigningAgent != "" {
		unsignedAttributes["signingAgent"] = envContent.SignerInfo.UnsignedAttributes.SigningAgent
	}

	return unsignedAttributes
}

func getCertificates(certChain []*x509.Certificate) []CertificateInfo {
	certificates := []CertificateInfo{}

	for _, cert := range certChain {
		h := sha256.Sum256(cert.Raw)
		fingerprint := strings.ToLower(hex.EncodeToString(h[:]))

		certificate := CertificateInfo{
			SHA256Fingerprint: fingerprint,
			IssuedTo:          cert.Subject.String(),
			IssuedBy:          cert.Issuer.String(),
			Expiry:            ioutil.Time(cert.NotAfter),
		}

		certificates = append(certificates, certificate)
	}

	return certificates
}

func parseTimestamp(signerInfo signature.SignerInfo) TimestampInfo {
	signedToken, err := tspclient.ParseSignedToken(signerInfo.UnsignedAttributes.TimestampSignature)
	if err != nil {
		return TimestampInfo{
			Error: fmt.Sprintf("failed to parse timestamp countersignature: %s", err.Error()),
		}
	}
	info, err := signedToken.Info()
	if err != nil {
		return TimestampInfo{
			Error: fmt.Sprintf("failed to parse timestamp countersignature: %s", err.Error()),
		}
	}
	timestamp, err := info.Validate(signerInfo.Signature)
	if err != nil {
		return TimestampInfo{
			Error: fmt.Sprintf("failed to parse timestamp countersignature: %s", err.Error()),
		}
	}
	return TimestampInfo{
		Timestamp:    ioutil.Timestamp(*timestamp),
		Certificates: getCertificates(signedToken.Certificates),
	}
}

// SignatureNode returns a tree node that represents the signature.
func (s *SignatureInfo) SignatureNode(sigName string) *tree.Node {
	sigNode := tree.New(sigName)
	sigNode.AddPair("signature algorithm", s.SignatureAlgorithm)
	sigNode.AddPair("signature envelope type", s.MediaType)

	signedAttributesNode := sigNode.Add("signed attributes")
	addMapToTree(signedAttributesNode, s.SignedAttributes)

	userDefinedAttributesNode := sigNode.Add("user defined attributes")
	addMapToTree(userDefinedAttributesNode, s.UserDefinedAttributes)

	unsignedAttributesNode := sigNode.Add("unsigned attributes")
	for k, v := range s.UnsignedAttributes {
		switch value := v.(type) {
		case string:
			unsignedAttributesNode.AddPair(k, value)
		case TimestampInfo:
			timestampNode := unsignedAttributesNode.Add("timestamp signature")
			if value.Error != "" {
				timestampNode.AddPair("error", value.Error)
				break
			}
			timestampNode.AddPair("timestamp", value.Timestamp)
			addCertificatesToTree(timestampNode, "certificates", value.Certificates)
		}
	}

	addCertificatesToTree(sigNode, "certificates", s.Certificates)

	artifactNode := sigNode.Add("signed artifact")
	artifactNode.AddPair("media type", s.SignedArtifact.MediaType)
	artifactNode.AddPair("digest", s.SignedArtifact.Digest.String())
	artifactNode.AddPair("size", strconv.FormatInt(s.SignedArtifact.Size, 10))
	return sigNode
}

func addMapToTree[T any](node *tree.Node, m map[string]T) {
	if len(m) > 0 {
		for k, v := range m {
			node.AddPair(k, v)
		}
	} else {
		node.Add("(empty)")
	}
}

func addCertificatesToTree(node *tree.Node, name string, certs []CertificateInfo) {
	certListNode := node.Add(name)
	for _, cert := range certs {
		certNode := certListNode.AddPair("SHA256 fingerprint", cert.SHA256Fingerprint)
		certNode.AddPair("issued to", cert.IssuedTo)
		certNode.AddPair("issued by", cert.IssuedBy)
		certNode.AddPair("expiry", cert.Expiry)
	}
}
