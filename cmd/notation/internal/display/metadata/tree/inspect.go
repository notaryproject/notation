package tree

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/cmd/notation/internal/display/metadata/model"
	"github.com/notaryproject/notation/cmd/notation/internal/output"
	"github.com/notaryproject/notation/internal/tree"
	"github.com/notaryproject/tspclient-go"
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

func (h *InspectHandler) AddSignature(digest string, envelopeMediaType string, sigEnvelope signature.Envelope) error {
	sig, err := model.NewSignature(digest, envelopeMediaType, sigEnvelope, formatter)
	if err != nil {
		return err
	}

	if h.root == nil || h.cncfSigNode == nil {
		return fmt.Errorf("artifact reference is not set")
	}

	sigNode := toTreeNode(sig)
	h.cncfSigNode.Children = append(h.cncfSigNode.Children, sigNode)
	return nil
}

func (h *InspectHandler) Print() error {
	if h.root == nil {
		return fmt.Errorf("artifact reference is not set")
	}
	return h.root.Print(h.printer)
}

func toTreeNode(s *model.Signature) *tree.Node {
	sigNode := tree.New(s.Digest)
	sigNode.AddPair("signature algorithm", s.SignatureAlgorithm)
	sigNode.AddPair("signature envelope type", s.MediaType)

	signedAttributesNode := sigNode.Add("signed attributes")
	addMapToTree(signedAttributesNode, s.SignedAttributes)

	userDefinedAttributesNode := sigNode.Add("user defined attributes")
	addMapToTree(userDefinedAttributesNode, s.UserDefinedAttributes)

	unsignedAttributesNode := sigNode.Add("unsigned attributes")
	for _, k := range orderedKeys(s.UnsignedAttributes) {
		v := s.UnsignedAttributes[k]
		switch value := v.(type) {
		case string:
			unsignedAttributesNode.AddPair(k, value)
		case model.Timestamp:
			timestampNode := unsignedAttributesNode.Add("timestamp signature")
			if value.Error != "" {
				timestampNode.AddPair("error", value.Error)
				continue
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

func addMapToTree(node *tree.Node, m map[string]string) {
	if len(m) == 0 {
		node.Add("(empty)")
		return
	}

	// Add each entry in sorted order
	for _, k := range orderedKeys(m) {
		node.AddPair(k, m[k])
	}
}

func orderedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func addCertificatesToTree(node *tree.Node, name string, certs []model.Certificate) {
	certListNode := node.Add(name)
	for _, cert := range certs {
		certNode := certListNode.AddPair("SHA256 fingerprint", cert.SHA256Fingerprint)
		certNode.AddPair("issued to", cert.IssuedTo)
		certNode.AddPair("issued by", cert.IssuedBy)
		certNode.AddPair("expiry", cert.Expiry)
	}
}

func formatter(v any) string {
	switch v := v.(type) {
	case time.Time:
		return v.Format(time.ANSIC)
	case tspclient.Timestamp:
		return v.Format(time.ANSIC)
	}
	return fmt.Sprint(v)
}
