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

package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation/internal/cmd"
	godigest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestInspectCommand_SecretsFromArgs(t *testing.T) {
	opts := &inspectOpts{}
	command := inspectCommand(opts)
	expected := &inspectOpts{
		reference: "ref",
		SecureFlagOpts: SecureFlagOpts{
			Password:         "password",
			InsecureRegistry: true,
			Username:         "user",
		},
		outputFormat:  cmd.OutputPlaintext,
		maxSignatures: 100,
	}
	if err := command.ParseFlags([]string{
		"--password", expected.Password,
		expected.reference,
		"-u", expected.Username,
		"--insecure-registry",
		"--output", "text"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if *opts != *expected {
		t.Fatalf("Expect inspect opts: %v, got: %v", expected, opts)
	}
}

func TestInspectCommand_SecretsFromEnv(t *testing.T) {
	t.Setenv(defaultUsernameEnv, "user")
	t.Setenv(defaultPasswordEnv, "password")
	opts := &inspectOpts{}
	expected := &inspectOpts{
		reference: "ref",
		SecureFlagOpts: SecureFlagOpts{
			Password: "password",
			Username: "user",
		},
		outputFormat:  cmd.OutputJSON,
		maxSignatures: 100,
	}
	command := inspectCommand(opts)
	if err := command.ParseFlags([]string{
		expected.reference,
		"--output", "json"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if *opts != *expected {
		t.Fatalf("Expect inspect opts: %v, got: %v", expected, opts)
	}
}

func TestInspectCommand_MissingArgs(t *testing.T) {
	command := inspectCommand(nil)
	if err := command.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}

func TestGetUnsignedAttributes(t *testing.T) {
	envContent := &signature.EnvelopeContent{
		SignerInfo: signature.SignerInfo{
			UnsignedAttributes: signature.UnsignedAttributes{
				TimestampSignature: []byte("invalid"),
			},
		},
	}
	expectedErrMsg := "failed to parse timestamp countersignature: cms: syntax error: invalid signed data: failed to convert from BER to DER: asn1: syntax error: decoding BER length octets: short form length octets value should be less or equal to the subsequent octets length"
	unsignedAttr := getUnsignedAttributes(cmd.OutputPlaintext, envContent)
	val, ok := unsignedAttr["timestampSignature"].(timestampOutput)
	if !ok {
		t.Fatal("expected to have timestampSignature")
	}
	if val.Error != expectedErrMsg {
		t.Fatalf("expected %s, but got %s", expectedErrMsg, val.Error)
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name         string
		outputFormat string
		signerInfo   signature.SignerInfo
		wantError    bool
	}{
		{
			name:         "invalid timestamp signature",
			outputFormat: cmd.OutputJSON,
			signerInfo: signature.SignerInfo{
				UnsignedAttributes: signature.UnsignedAttributes{
					TimestampSignature: []byte("invalid"),
				},
			},
			wantError: true,
		},
		{
			name:         "empty timestamp signature",
			outputFormat: cmd.OutputJSON,
			signerInfo: signature.SignerInfo{
				UnsignedAttributes: signature.UnsignedAttributes{
					TimestampSignature: []byte{},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTimestamp(tt.outputFormat, tt.signerInfo)
			if tt.wantError && result.Error == "" {
				t.Error("expected error but got none")
			}
			if !tt.wantError && result.Error != "" {
				t.Errorf("unexpected error: %s", result.Error)
			}
		})
	}
}

func tearDown(td struct {
	testTime  time.Time
	testRef   string
	testSig   signatureOutput
	output    inspectOutput
	oldStdout *os.File
	r, w      *os.File
	buf       *bytes.Buffer
}) {
	td.w.Close()
	os.Stdout = td.oldStdout
}

func getOutput(td struct {
	testTime  time.Time
	testRef   string
	testSig   signatureOutput
	output    inspectOutput
	oldStdout *os.File
	r, w      *os.File
	buf       *bytes.Buffer
}) string {
	td.w.Close()
	td.buf.Reset()
	_, _ = td.buf.ReadFrom(td.r)
	return td.buf.String()
}

func TestPrintOutput_InvalidFormat(t *testing.T) {
	td := setupTest(t)
	defer tearDown(td)

	err := printOutput("invalid-format", td.testRef, td.output)
	if err == nil {
		t.Error("expected error for invalid format, got nil")
	}
	if err.Error() != "unrecognized output format invalid-format" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestPrintOutput_PlaintextFormat tests the plaintext output
func TestPrintOutput_PlaintextFormat(t *testing.T) {
	td := setupTest(t)
	defer tearDown(td)

	// Convert times to RFC3339 format for plaintext output
	sig1Time, _ := time.Parse(time.RFC3339, td.output.Signatures[0].SignedAttributes["signingTime"])
	sig1Expiry, _ := time.Parse(time.RFC3339, td.output.Signatures[0].Certificates[0].Expiry)
	sig2Time, _ := time.Parse(time.RFC3339, td.output.Signatures[1].SignedAttributes["signingTime"])
	sig2Expiry, _ := time.Parse(time.RFC3339, td.output.Signatures[1].Certificates[0].Expiry)
	td.output.Signatures[0].SignedAttributes["signingTime"] = sig1Time.Format(time.RFC3339)
	td.output.Signatures[0].Certificates[0].Expiry = sig1Expiry.Format(time.RFC3339)
	td.output.Signatures[1].SignedAttributes["signingTime"] = sig2Time.Format(time.RFC3339)
	td.output.Signatures[1].Certificates[0].Expiry = sig2Expiry.Format(time.RFC3339)

	err := printOutput("text", td.testRef, td.output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outputStr := getOutput(td)

	expectedOutput := `Inspecting all signatures for signed artifact
docker.io/holiodin01/net-monitor@sha256:984a85875f80346ec12aab68fefa1d816dd3916175cf18938f498d2f0677c208
└── application/vnd.cncf.notary.signature
    ├── sha256:efaea8b26ef007bfbdaecab63b068320e508921b7687e38b892fd4ce71a92f4e
    │   ├── media type: application/jose+json
    │   ├── signature algorithm: RSASSA-PSS-SHA-256
    │   ├── signed attributes
    │   │   ├── signingScheme: notary.x509
    │   │   └── signingTime: 2024-12-28T11:11:16+05:30
    │   ├── user defined attributes
    │   │   └── (empty)
    │   ├── unsigned attributes
    │   │   └── signingAgent: notation-go/1.3.0+unreleased
    │   ├── certificates
    │   │   └── SHA256 fingerprint: 1487c63cf13ae92488c2bd3d9bcb95687c4bb8e90d501a2f15956e42e47ba06e
    │   │       ├── issued to: CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US
    │   │       ├── issued by: CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US
    │   │       └── expiry: 2024-12-29T05:40:12Z
    │   └── signed artifact
    │       ├── media type: application/vnd.docker.distribution.manifest.v2+json
    │       ├── digest: sha256:984a85875f80346ec12aab68fefa1d816dd3916175cf18938f498d2f0677c208
    │       └── size: 942
    └── sha256:7612778026574769361c778fa8c7b8935f273ad85425b637561167c6629ce369
        ├── media type: application/jose+json
        ├── signature algorithm: RSASSA-PSS-SHA-256
        ├── signed attributes
        │   ├── signingScheme: notary.x509
        │   └── signingTime: 2024-12-28T11:13:47+05:30
        ├── user defined attributes
        │   └── (empty)
        ├── unsigned attributes
        │   └── signingAgent: notation-go/1.3.0+unreleased
        ├── certificates
        │   └── SHA256 fingerprint: 1487c63cf13ae92488c2bd3d9bcb95687c4bb8e90d501a2f15956e42e47ba06e
        │       ├── issued to: CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US
        │       ├── issued by: CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US
        │       └── expiry: 2024-12-29T05:40:12Z
        └── signed artifact
            ├── media type: application/vnd.docker.distribution.manifest.v2+json
            ├── digest: sha256:984a85875f80346ec12aab68fefa1d816dd3916175cf18938f498d2f0677c208
            └── size: 942`

	expectedLines := strings.Split(expectedOutput, "\n")
	outputLines := strings.Split(outputStr, "\n")

	// Normalize the indentation level and strip unnecessary spaces for comparison
	for i := range expectedLines {
		expectedLines[i] = normalizeLine(expectedLines[i])
	}
	for i := range outputLines {
		outputLines[i] = normalizeLine(outputLines[i])
	}

	// Compare the lines of output
	for i := range expectedLines {
		if i >= len(outputLines) {
			t.Errorf("Output is missing a line: %s", expectedLines[i])
			continue
		}
		if expectedLines[i] != outputLines[i] {
			t.Errorf("Line %d does not match:\nExpected: %s\nGot: %s", i+1, expectedLines[i], outputLines[i])
		}
	}

	// Check if the actual output has extra lines
	if len(outputLines) > len(expectedLines) {
		for i := len(expectedLines); i < len(outputLines); i++ {
			t.Logf("Unexpected extra line: %s", outputLines[i])
		}
	}
}

// normalizeLine trims leading/trailing spaces and normalizes indentation
func normalizeLine(line string) string {
	line = strings.TrimSpace(line)
	// Normalize indentation, i.e., replace all leading spaces with a consistent indentation
	return line
}

func setupTest(t *testing.T) (testData struct {
	testTime  time.Time
	testRef   string
	testSig   signatureOutput
	output    inspectOutput
	oldStdout *os.File
	r, w      *os.File
	buf       *bytes.Buffer
}) {
	artifactDigest := "sha256:984a85875f80346ec12aab68fefa1d816dd3916175cf18938f498d2f0677c208"
	testData.testRef = "docker.io/holiodin01/net-monitor@" + artifactDigest

	// Correct the format of the signingTime to match the expected format
	signingTime1 := "2024-12-28T11:11:16+05:30"
	signingTime2 := "2024-12-28T11:13:47+05:30"

	sig1 := signatureOutput{
		MediaType:          "application/jose+json",
		Digest:             "sha256:efaea8b26ef007bfbdaecab63b068320e508921b7687e38b892fd4ce71a92f4e",
		SignatureAlgorithm: "RSASSA-PSS-SHA-256",
		SignedAttributes: map[string]string{
			"signingScheme": "notary.x509",
			"signingTime":   signingTime1,
		},
		UserDefinedAttributes: map[string]string{},
		UnsignedAttributes: map[string]any{
			"signingAgent": "notation-go/1.3.0+unreleased",
		},
		Certificates: []certificateOutput{
			{
				SHA256Fingerprint: "1487c63cf13ae92488c2bd3d9bcb95687c4bb8e90d501a2f15956e42e47ba06e",
				IssuedTo:          "CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US",
				IssuedBy:          "CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US",
				Expiry:            "2024-12-29T05:40:12Z",
			},
		},
		SignedArtifact: ocispec.Descriptor{
			MediaType: "application/vnd.docker.distribution.manifest.v2+json",
			Digest:    godigest.Digest(artifactDigest),
			Size:      942,
		},
	}

	sig2 := signatureOutput{
		MediaType:          "application/jose+json",
		Digest:             "sha256:7612778026574769361c778fa8c7b8935f273ad85425b637561167c6629ce369",
		SignatureAlgorithm: "RSASSA-PSS-SHA-256",
		SignedAttributes: map[string]string{
			"signingScheme": "notary.x509",
			"signingTime":   signingTime2,
		},
		UserDefinedAttributes: map[string]string{},
		UnsignedAttributes: map[string]any{
			"signingAgent": "notation-go/1.3.0+unreleased",
		},
		Certificates: []certificateOutput{
			{
				SHA256Fingerprint: "1487c63cf13ae92488c2bd3d9bcb95687c4bb8e90d501a2f15956e42e47ba06e",
				IssuedTo:          "CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US",
				IssuedBy:          "CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US",
				Expiry:            "2024-12-29T05:40:12Z",
			},
		},
		SignedArtifact: ocispec.Descriptor{
			MediaType: "application/vnd.docker.distribution.manifest.v2+json",
			Digest:    godigest.Digest(artifactDigest),
			Size:      942,
		},
	}

	testData.output = inspectOutput{
		MediaType:  "application/vnd.docker.distribution.manifest.v2+json",
		Signatures: []signatureOutput{sig1, sig2},
	}

	// Capture stdout
	testData.oldStdout = os.Stdout
	var err error
	testData.r, testData.w, err = os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = testData.w
	testData.buf = &bytes.Buffer{}

	return testData
}

func TestPrintOutput_JSONFormat(t *testing.T) {
	td := setupTest(t)
	// Convert times to RFC3339 format for JSON output!
	sig1Time, _ := time.Parse(time.RFC3339, td.output.Signatures[0].SignedAttributes["signingTime"])
	sig1Expiry, _ := time.Parse(time.RFC3339, td.output.Signatures[0].Certificates[0].Expiry)
	sig2Time, _ := time.Parse(time.RFC3339, td.output.Signatures[1].SignedAttributes["signingTime"])
	sig2Expiry, _ := time.Parse(time.RFC3339, td.output.Signatures[1].Certificates[0].Expiry)

	td.output.Signatures[0].SignedAttributes["signingTime"] = sig1Time.Format(time.RFC3339)
	td.output.Signatures[0].Certificates[0].Expiry = sig1Expiry.Format(time.RFC3339)
	td.output.Signatures[1].SignedAttributes["signingTime"] = sig2Time.Format(time.RFC3339)
	td.output.Signatures[1].Certificates[0].Expiry = sig2Expiry.Format(time.RFC3339)

	td.output.Signatures[0].UserDefinedAttributes = nil
	td.output.Signatures[1].UserDefinedAttributes = nil

	err := printOutput("json", td.testRef, td.output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outputStr := getOutput(td)

	expectedOutput := `{
    "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
    "Signatures": [
        {
            "mediaType": "application/jose+json",
            "digest": "sha256:efaea8b26ef007bfbdaecab63b068320e508921b7687e38b892fd4ce71a92f4e",
            "signatureAlgorithm": "RSASSA-PSS-SHA-256",
            "signedAttributes": {
                "signingScheme": "notary.x509",
                "signingTime": "2024-12-28T11:11:16+05:30"
            },
            "userDefinedAttributes": null,
            "unsignedAttributes": {
                "signingAgent": "notation-go/1.3.0+unreleased"
            },
            "certificates": [
                {
                    "SHA256Fingerprint": "1487c63cf13ae92488c2bd3d9bcb95687c4bb8e90d501a2f15956e42e47ba06e",
                    "issuedTo": "CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US",
                    "issuedBy": "CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US",
                    "expiry": "2024-12-29T05:40:12Z"
                }
            ],
            "signedArtifact": {
                "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
                "digest": "sha256:984a85875f80346ec12aab68fefa1d816dd3916175cf18938f498d2f0677c208",
                "size": 942
            }
        },
        {
            "mediaType": "application/jose+json",
            "digest": "sha256:7612778026574769361c778fa8c7b8935f273ad85425b637561167c6629ce369",
            "signatureAlgorithm": "RSASSA-PSS-SHA-256",
            "signedAttributes": {
                "signingScheme": "notary.x509",
                "signingTime": "2024-12-28T11:13:47+05:30"
            },
            "userDefinedAttributes": null,
            "unsignedAttributes": {
                "signingAgent": "notation-go/1.3.0+unreleased"
            },
            "certificates": [
                {
                    "SHA256Fingerprint": "1487c63cf13ae92488c2bd3d9bcb95687c4bb8e90d501a2f15956e42e47ba06e",
                    "issuedTo": "CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US",
                    "issuedBy": "CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US",
                    "expiry": "2024-12-29T05:40:12Z"
                }
            ],
            "signedArtifact": {
                "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
                "digest": "sha256:984a85875f80346ec12aab68fefa1d816dd3916175cf18938f498d2f0677c208",
                "size": 942
            }
        }
    ]
}`

	expectedLines := strings.Split(expectedOutput, "\n")
	outputLines := strings.Split(outputStr, "\n")

	for i := range expectedLines {
		if i >= len(outputLines) {
			t.Errorf("Output is missing a line: %s", expectedLines[i])
			continue
		}
		if expectedLines[i] != outputLines[i] {
			t.Errorf("Line %d does not match:\nExpected: %s\nGot: %s", i+1, expectedLines[i], outputLines[i])
		}
	}

	// Check if the actual output has extra lines (this will break the test if the expected output is missing lines)
	if len(outputLines) > len(expectedLines) {
		for i := len(expectedLines); i < len(outputLines); i++ {
			t.Errorf("Unexpected extra line: %s", outputLines[i])
		}
	}
}
