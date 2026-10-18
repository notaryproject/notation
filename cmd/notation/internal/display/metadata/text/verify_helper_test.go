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
	"maps"
	"slices"
	"strings"
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

func TestUserMetadataIsSortedByKey(t *testing.T) {
	annotations := map[string]string{
		"org.example.zone":    "us-east",
		"org.example.alpha":   "1",
		"io.wabbit.buildTime": "2024-01-01",
		"io.wabbit.buildId":   "123",
		"com.example.team":    "platform",
		"a":                   "first",
		"z":                   "last",
		"m":                   "middle",
	}
	payloadBytes, err := json.Marshal(&envelope.Payload{
		TargetArtifact: ocispec.Descriptor{Annotations: annotations},
	})
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	outcome := &notation.VerificationOutcome{
		EnvelopeContent: &signature.EnvelopeContent{
			Payload: signature.Payload{Content: payloadBytes},
		},
	}

	want := slices.Sorted(maps.Keys(annotations))
	// Go randomizes map iteration order, so a single run can be sorted by
	// chance. Render repeatedly to make the check meaningful.
	for i := 0; i < 20; i++ {
		buf := bytes.Buffer{}
		h := NewVerifyHandler(output.NewPrinter(&buf, &buf))
		h.OnVerifySucceeded([]*notation.VerificationOutcome{outcome}, "localhost:5000/net-monitor@sha256:abcd")
		if err := h.Render(); err != nil {
			t.Fatalf("Render() error: %v", err)
		}
		got := metadataKeys(t, buf.String())
		if !slices.Equal(got, want) {
			t.Fatalf("run %d: user metadata keys are not sorted: got %v, want %v", i, got, want)
		}
	}
}

// metadataKeys returns the key column of the user metadata table printed by
// the verify handler. It fails the test when the table is missing so that an
// empty result cannot be mistaken for sorted output.
func metadataKeys(t *testing.T, out string) []string {
	t.Helper()
	if !strings.Contains(out, "Successfully verified signature for") {
		t.Fatalf("verification success message missing from output: %q", out)
	}
	lines := strings.Split(out, "\n")
	header := slices.IndexFunc(lines, func(line string) bool {
		return strings.HasPrefix(line, "KEY")
	})
	if header == -1 {
		t.Fatalf("user metadata table missing from output: %q", out)
	}
	var keys []string
	for _, line := range lines[header+1:] {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		keys = append(keys, fields[0])
	}
	return keys
}
