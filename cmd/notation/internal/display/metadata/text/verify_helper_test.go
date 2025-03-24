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
	"bytes"
	"encoding/json"
	"testing"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/v2/cmd/notation/internal/display/output"
	"github.com/notaryproject/notation/v2/internal/envelope"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestPrintMetadataIfPresent(t *testing.T) {
	payload := &envelope.Payload{
		TargetArtifact: ocispec.Descriptor{
			Annotations: map[string]string{
				"foo": "bar",
			},
		},
	}
	payloadBytes, _ := json.Marshal(payload)

	outcome := &notation.VerificationOutcome{
		EnvelopeContent: &signature.EnvelopeContent{
			Payload: signature.Payload{
				Content: payloadBytes,
			},
		},
	}

	t.Run("with metadata", func(t *testing.T) {
		buf := bytes.Buffer{}
		printer := output.NewPrinter(&buf, &buf)
		h := NewVerifyHandler(printer)
		printUserMetadataIfPresent(h.printer, outcome)
		got := buf.String()
		expected := "\nThe artifact was signed with the following user metadata.\n\nKEY   VALUE   \nfoo   bar     \n"
		if got != expected {
			t.Errorf("unexpected output: %q", got)
		}
	})
}
