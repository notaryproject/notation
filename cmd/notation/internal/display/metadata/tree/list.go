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
	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/cmd/notation/internal/display/output"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// ListHandler is a handler for rendering signature metadata information in
// a tree format. It implements the ListHandler interface.
type ListHandler struct {
	printer        *output.Printer
	root           *node
	signaturesNode *node
}

// NewListHandler creates a new ListHandler.
func NewListHandler(printer *output.Printer) *ListHandler {
	return &ListHandler{
		printer: printer,
	}
}

// OnResolvingTagReference outputs the tag reference warning.
func (h *ListHandler) OnResolvingTagReference(reference string) {
	h.printer.PrintErrorf("Warning: Always list the artifact using digest(@sha256:...) rather than a tag(:%s) because resolved digest may not point to the same signed artifact, as tags are mutable.\n", reference)
}

// OnReferenceResolved sets the artifact reference and media type for the
// handler.
func (h *ListHandler) OnReferenceResolved(reference string) {
	h.root = newNode(reference)
	h.signaturesNode = h.root.Add(notationregistry.ArtifactTypeNotation)
}

// ListSignature adds the signature digest to the tree.
func (h *ListHandler) ListSignature(signatureManifest ocispec.Descriptor) {
	h.signaturesNode.Add(signatureManifest.Digest.String())
}

// OnExceedMaxSignatures outputs the warning message when the number of
// signatures exceeds the maximum limit.
func (h *ListHandler) OnExceedMaxSignatures(err error) {
	h.printer.PrintErrorf("Warning: %v\n", err)
}

// Render prints the tree format of the signature metadata information.
func (h *ListHandler) Render() error {
	if h.root == nil || h.signaturesNode == nil || len(h.signaturesNode.Children) == 0 {
		return h.printer.Printf("%s has no associated signatures\n", h.root.Value)
	}
	return h.root.Print(h.printer)
}
