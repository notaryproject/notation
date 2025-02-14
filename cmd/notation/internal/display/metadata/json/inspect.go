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
	coresignature "github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation/cmd/notation/internal/display/output"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type inspectOutput struct {
	MediaType  string       `json:"mediaType"`
	Signatures []*signature `json:"signatures"`
}

// InspectHandler is the handler for inspecting metadata information and
// rendering it in JSON format. It implements the metadata.InspectHandler
// interface.
type InspectHandler struct {
	printer *output.Printer

	output inspectOutput
}

// NewInspectHandler creates an Inspecthandler to inspect signatures and print in
// JSON format.
func NewInspectHandler(printer *output.Printer) *InspectHandler {
	return &InspectHandler{
		printer: printer,
		output: inspectOutput{
			Signatures: []*signature{},
		},
	}
}

// OnReferenceResolved sets the artifact reference and media type for the
// handler.
//
// The reference is no-op for this handler.
func (h *InspectHandler) OnReferenceResolved(_, mediaType string) {
	h.output.MediaType = mediaType
}

// InspectSignature inspects a signature to get it ready to be rendered.
func (h *InspectHandler) InspectSignature(manifestDesc, signatureDesc ocispec.Descriptor, envelope coresignature.Envelope) error {
	sig, err := newSignature(manifestDesc.Digest.String(), signatureDesc.MediaType, envelope)
	if err != nil {
		return err
	}
	h.output.Signatures = append(h.output.Signatures, sig)
	return nil
}

// Render renders signatures metadata information in JSON format.
func (h *InspectHandler) Render() error {
	return output.PrintPrettyJSON(h.printer, h.output)
}
