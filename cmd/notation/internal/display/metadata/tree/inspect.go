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
	coresignature "github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/cmd/notation/internal/display/output"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// InspectHandler is a handler for inspecting metadata information and rendering
// it in a tree format. It implements the metadata.InspectHandler interface.
type InspectHandler struct {
	printer *output.Printer

	// rootReferenceNode is the root node with the artifact reference as the
	// value.
	rootReferenceNode *node
	// notationSignaturesNode is the node for all signatures associated with the
	// artifact.
	notationSignaturesNode *node
}

// NewInspectHandler creates a InspectHandler to inspect signature and print in
// tree format.
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
	h.rootReferenceNode = newNode(reference)
	h.notationSignaturesNode = h.rootReferenceNode.Add(registry.ArtifactTypeNotation)
}

// InspectSignature inspects a signature to get it ready to be rendered.
func (h *InspectHandler) InspectSignature(manifestDesc, signatureDesc ocispec.Descriptor, envelope coresignature.Envelope) error {
	sigNode, err := newSignatureNode(manifestDesc.Digest.String(), signatureDesc.MediaType, envelope)
	if err != nil {
		return err
	}
	h.notationSignaturesNode.Children = append(h.notationSignaturesNode.Children, sigNode)
	return nil
}

// Render renders the metadata information when an operation is complete.
func (h *InspectHandler) Render() error {
	if len(h.notationSignaturesNode.Children) == 0 {
		return h.printer.Printf("%s has no associated signature\n", h.rootReferenceNode.Value)
	}
	h.printer.Println("Inspecting all signatures for signed artifact")
	return h.rootReferenceNode.Print(h.printer)
}
