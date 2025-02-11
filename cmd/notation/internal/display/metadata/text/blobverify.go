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

// BlobVerifyHandler is a handler for rendering output for blob verify command
// in human-readable format.
type BlobVerifyHandler struct {
	printer *output.Printer

	outcome  *notation.VerificationOutcome
	blobPath string
}

// NewBlobVerifyHandler creates a new BlobVerifyHandler.
func NewBlobVerifyHandler(printer *output.Printer) *BlobVerifyHandler {
	return &BlobVerifyHandler{
		printer: printer,
	}
}

// OnVerifySucceeded sets the successful verification result for the handler.
//
// outcomes must not be nil or empty.
func (h *BlobVerifyHandler) OnVerifySucceeded(outcomes []*notation.VerificationOutcome, blobPath string) {
	h.outcome = outcomes[0]
	h.blobPath = blobPath
}

// Render prints out the verification results in human-readable format.
func (h *BlobVerifyHandler) Render() error {
	return PrintVerificationSuccess(h.printer, h.outcome, h.blobPath, false)
}
