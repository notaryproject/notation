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
	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation/cmd/notation/internal/display/output"
)

// BlobInspectHandler is a handler for inspecting metadata information and
// rendering it in a tree format. It implements the metadata.BlobInspectHandler
// interface.
type BlobInspectHandler struct {
	printer *output.Printer

	signatureNode *node
}

// NewBlobInspectHandler creates a BlobInspectHandler to inspect signature and
// print in tree format.
func NewBlobInspectHandler(printer *output.Printer) *BlobInspectHandler {
	return &BlobInspectHandler{
		printer: printer,
	}
}

// OnEnvelopeParsed sets the parsed envelope for the handler.
func (h *BlobInspectHandler) OnEnvelopeParsed(nodeName, envelopeMediaType string, envelope signature.Envelope) error {
	sigNode, err := newSignatureNode(nodeName, envelopeMediaType, envelope)
	if err != nil {
		return err
	}
	h.signatureNode = sigNode
	return nil
}

// Render prints out the metadata information in tree format.
func (h *BlobInspectHandler) Render() error {
	return h.signatureNode.Print(h.printer)
}
