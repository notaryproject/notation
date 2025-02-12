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

package text

import (
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/cmd/notation/internal/display/output"
)

// VerifyHandler is a handler for rendering output for verify command in
// human-readable format.
type VerifyHandler struct {
	printer         *output.Printer
	outcome         *notation.VerificationOutcome
	digestReference string
	hasWarning      bool
}

// NewVerifyHandler creates a new VerifyHandler.
func NewVerifyHandler(printer *output.Printer) *VerifyHandler {
	return &VerifyHandler{
		printer: printer,
	}
}

// OnResolvingTagReference outputs the tag reference warning.
func (h *VerifyHandler) OnResolvingTagReference(reference string) {
	h.printer.PrintErrorf("Warning: Always verify the artifact using digest(@sha256:...) rather than a tag(:%s) because resolved digest may not point to the same signed artifact, as tags are mutable.\n", reference)
	h.hasWarning = true
}

// OnVerifySucceeded sets the successful verification result for the handler.
//
// outcomes must not be nil or empty.
func (h *VerifyHandler) OnVerifySucceeded(outcomes []*notation.VerificationOutcome, digestReference string) {
	h.outcome = outcomes[0]
	h.digestReference = digestReference
}

// Render prints out the verification results in human-readable format.
func (h *VerifyHandler) Render() error {
	return PrintVerificationSuccess(h.printer, h.outcome, h.digestReference, h.hasWarning)
}
