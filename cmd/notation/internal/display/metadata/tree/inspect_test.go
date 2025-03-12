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
	"errors"
	"testing"

	coresignature "github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation/v2/cmd/notation/internal/display/output"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// errorWriter is a mock io.Writer that always returns an error
type errorWriter struct{}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("mocked write error")
}

func TestInspectHandler_InspectSignature_Print_Error(t *testing.T) {
	// Create a printer with an error writer
	errorPrinter := output.NewPrinter(&errorWriter{}, &errorWriter{})
	handler := NewInspectHandler(errorPrinter)

	// Set reference to initialize rootReferenceNode
	handler.OnReferenceResolved("test-reference", "")

	// Mock descriptor and envelope
	manifestDesc := ocispec.Descriptor{
		Digest: "sha256:test",
	}
	signatureDesc := ocispec.Descriptor{
		MediaType: "test-media-type",
	}
	var envelope coresignature.Envelope

	// Test InspectSignature - should fail at printing header
	err := handler.InspectSignature(manifestDesc, signatureDesc, envelope)
	if err == nil {
		t.Fatal("Expected error when printing header, but got nil")
	}
}

func TestInspectHandler_Render_Flush_Error(t *testing.T) {
	// Create a printer with an error writer
	errorPrinter := output.NewPrinter(&errorWriter{}, &errorWriter{})
	handler := NewInspectHandler(errorPrinter)

	// Add a signature to ensure sprinter has something to flush
	handler.OnReferenceResolved("test-reference", "")
	handler.headerPrinted = true // Simulate header already printed

	handler.sprinter.prevNode = &node{} // Simulate a node to flush

	// Test Render - should fail at flushing
	err := handler.Render()
	if err == nil {
		t.Fatal("Expected error when flushing, but got nil")
	}
}
