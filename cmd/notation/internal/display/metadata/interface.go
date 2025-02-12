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

// Package metadata defines interfaces for handlers that render metadata
// information for each command. The metadata provides information about the
// original data with formatted output in JSON, tree, or text.
package metadata

import (
	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Renderer renders metadata information when an operation is complete.
type Renderer interface {
	Render() error
}

// InspectHandler is a handler for inspecting metadata information and rendering
// it in a specific format.
type InspectHandler interface {
	Renderer

	// OnReferenceResolved sets the artifact reference and media type for the handler.
	OnReferenceResolved(reference, mediaType string)

	// InspectSignature inspects a signature to get it ready to be rendered.
	InspectSignature(manifestDesc, signatureDesc ocispec.Descriptor, envelope signature.Envelope) error
}

// BlobInspectHandler is a handler for rendering metadata information of a blob
// signature.
type BlobInspectHandler interface {
	Renderer

	// OnEvelopeParsed sets the parsed envelope for the handler.
	OnEnvelopeParsed(nodeName, envelopeMediaType string, envelope signature.Envelope) error
}

// VerifyHandler is a handler for rendering metadata information of
// verification outcome.
//
// It only supports text format for now.
type VerifyHandler interface {
	Renderer

	// OnResolvingTagReference outputs the tag reference warning.
	OnResolvingTagReference(reference string)

	// OnVerifySucceeded sets the successful verification result for the handler.
	//
	// outcomes must not be nil or empty.
	OnVerifySucceeded(outcomes []*notation.VerificationOutcome, digestReference string)
}
