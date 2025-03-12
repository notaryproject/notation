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
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	coresignature "github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go/plugin/proto"
	envelopeutil "github.com/notaryproject/notation/v2/internal/envelope"
	"github.com/notaryproject/tspclient-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func newSignatureNode(nodeName, signatureMediaType string, envelope coresignature.Envelope) (*node, error) {
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

	// create signature node
	sigNode := newNode(nodeName)
	sigNode.AddPair("signature algorithm", string(signatureAlgorithm))
	sigNode.AddPair("signature envelope type", signatureMediaType)

	addSignedAttributes(sigNode, envelopeContent)
	addUserDefinedAttributes(sigNode, signedArtifactDesc.Annotations)
	addUnsignedAttributes(sigNode, envelopeContent)
	addCertificates(sigNode, envelopeContent.SignerInfo.CertificateChain)
	addSignedArtifact(sigNode, signedArtifactDesc)
	return sigNode, nil
}

func addSignedAttributes(node *node, envelopeContent *coresignature.EnvelopeContent) {
	signedAttributesNode := node.Add("signed attributes")
	signedAttributesNode.AddPair("content type", string(envelopeContent.Payload.ContentType))
	signedAttributesNode.AddPair("signing scheme", string(envelopeContent.SignerInfo.SignedAttributes.SigningScheme))
	signedAttributesNode.AddPair("signing time", formatTime(envelopeContent.SignerInfo.SignedAttributes.SigningTime))
	if expiry := envelopeContent.SignerInfo.SignedAttributes.Expiry; !expiry.IsZero() {
		signedAttributesNode.AddPair("expiry", formatTime(expiry))
	}
	for _, attribute := range envelopeContent.SignerInfo.SignedAttributes.ExtendedAttributes {
		signedAttributesNode.AddPair(fmt.Sprint(attribute.Key), fmt.Sprint(attribute.Value))
	}
}

func addUserDefinedAttributes(node *node, annotations map[string]string) {
	userDefinedAttributesNode := node.Add("user defined attributes")
	if len(annotations) == 0 {
		userDefinedAttributesNode.Add("(empty)")
		return
	}
	for _, k := range slices.Sorted(maps.Keys(annotations)) {
		v := annotations[k]
		userDefinedAttributesNode.AddPair(k, v)
	}
}

func addUnsignedAttributes(node *node, envelopeContent *coresignature.EnvelopeContent) {
	unsignedAttributesNode := node.Add("unsigned attributes")
	if signingAgent := envelopeContent.SignerInfo.UnsignedAttributes.SigningAgent; signingAgent != "" {
		unsignedAttributesNode.AddPair("signing agent", signingAgent)
	}
	if timestamp := envelopeContent.SignerInfo.UnsignedAttributes.TimestampSignature; timestamp != nil {
		addTimestamp(unsignedAttributesNode, envelopeContent.SignerInfo)
	}
}

func addSignedArtifact(node *node, signedArtifactDesc ocispec.Descriptor) {
	artifactNode := node.Add("signed artifact")
	artifactNode.AddPair("media type", signedArtifactDesc.MediaType)
	artifactNode.AddPair("digest", signedArtifactDesc.Digest.String())
	artifactNode.AddPair("size", strconv.FormatInt(signedArtifactDesc.Size, 10))
}

func addTimestamp(node *node, signerInfo coresignature.SignerInfo) {
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

func addCertificates(node *node, certChain []*x509.Certificate) {
	certListNode := node.Add("certificates")
	for _, cert := range certChain {
		hash := sha256.Sum256(cert.Raw)
		certNode := certListNode.AddPair("SHA256 fingerprint", strings.ToLower(hex.EncodeToString(hash[:])))
		certNode.AddPair("issued to", cert.Subject.String())
		certNode.AddPair("issued by", cert.Issuer.String())
		certNode.AddPair("expiry", formatTime(cert.NotAfter))
	}
}

func formatTime(t time.Time) string {
	return t.Format(time.ANSIC)
}
