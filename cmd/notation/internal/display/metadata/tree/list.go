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

// ListHandler is a handler for rendering a list of signature digests in
// streaming fashion. It implements the metadata.ListHandler interface.
type ListHandler struct {
	printer *output.Printer

	// sprinter is a streaming printer to print the signature digest nodes in
	// a streaming fashion
	sprinter *streamingPrinter

	// headerNode contains the headers of the output
	//
	// example:
	// localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac4efe37a5380ee9088f7ace2efcde9
	// └── application/vnd.cncf.notary.signature
	headerNode *node

	// headerPrinted is a flag to indicate if the header has been printed
	headerPrinted bool
}

// NewListHandler creates a new ListHandler.
func NewListHandler(printer *output.Printer) *ListHandler {
	return &ListHandler{
		printer:  printer,
		sprinter: newStreamingPrinter("    ", printer),
	}
}

// OnResolvingTagReference outputs the tag reference warning.
func (h *ListHandler) OnResolvingTagReference(reference string) {
	h.printer.PrintErrorf("Warning: Always list the artifact using digest(@sha256:...) rather than a tag(:%s) because resolved digest may not point to the same signed artifact, as tags are mutable.\n", reference)
}

// OnReferenceResolved sets the artifact reference and media type for the
// handler.
func (h *ListHandler) OnReferenceResolved(reference string) {
	h.headerNode = newNode(reference)
	h.headerNode.Add(notationregistry.ArtifactTypeNotation)
}

// OnSignatureListed adds the signature digest to be printed.
func (h *ListHandler) OnSignatureListed(signatureManifest ocispec.Descriptor) {
	// print the header
	if !h.headerPrinted {
		h.headerNode.Print(h.printer)
		h.headerPrinted = true
	}
	h.sprinter.PrintNode(newNode(signatureManifest.Digest.String()))
}

// OnExceedMaxSignatures outputs the warning message when the number of
// signatures exceeds the maximum limit.
func (h *ListHandler) OnExceedMaxSignatures(err error) {
	h.printer.PrintErrorf("Warning: %v\n", err)
}

// Render completes the rendering of the list of signature digests.
func (h *ListHandler) Render() error {
	if h.sprinter.prevNode == nil {
		return h.printer.Printf("%s has no associated signatures\n", h.headerNode.Value)
	}
	h.sprinter.Complete()
	return nil
}
