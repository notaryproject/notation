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

package metadata

import "github.com/notaryproject/notation-core-go/signature"

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
	InspectSignature(digest string, envelopeMediaType string, sigEnvelope signature.Envelope) error
}
