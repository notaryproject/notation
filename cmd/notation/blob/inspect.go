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

package blob

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/notaryproject/notation-core-go/signature/cose"
	"github.com/notaryproject/notation-core-go/signature/jws"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/notaryproject/notation/internal/tree"
	"github.com/spf13/cobra"
)

type inspectOpts struct {
	sigPath      string
	outputFormat string
}

func inspectCommand() *cobra.Command {
	opts := &inspectOpts{}
	command := &cobra.Command{
		Use:   "inspect [flags] <signature_path>",
		Short: "Inspect a signature associated with a blob",
		Long: `Inspect a signature associated with a blob.

Example - Inspect a signature:
  notation inspect blob.cose.sig
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing signature path: use `notation blob inspect --help` to see what parameters are required")
			}
			opts.sigPath = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInspect(opts)
		},
	}

	cmd.SetPflagOutput(command.Flags(), &opts.outputFormat, cmd.PflagOutputUsage)
	return command
}

func runInspect(opts *inspectOpts) error {
	if opts.outputFormat != cmd.OutputJSON && opts.outputFormat != cmd.OutputPlaintext {
		return fmt.Errorf("unrecognized output format %s", opts.outputFormat)
	}

	envelopeMediaType, err := parseEnvelopeMediaType(filepath.Base(opts.sigPath))
	if err != nil {
		return err
	}

	sigBlob, err := os.ReadFile(opts.sigPath)
	if err != nil {
		return err
	}

	sig, err := envelope.Parse(sigBlob, envelopeMediaType)
	if err != nil {
		return err
	}

	// displayed as UserDefinedAttributes
	sig.SignedArtifact.Annotations = nil

	return printOutput(opts.sigPath, sig, opts.outputFormat)
}

func printOutput(sigPath string, signature *envelope.SignatureInfo, outputFormat string) error {
	if outputFormat == cmd.OutputJSON {
		return ioutil.PrintObjectAsJSON(signature)
	}

	sigNode := tree.New(sigPath)
	sigNode.AddPair("signature algorithm", signature.SignatureAlgorithm)
	sigNode.AddPair("signature envelope type", signature.MediaType)

	signedAttributesNode := sigNode.Add("signed attributes")
	addMapToTree(signedAttributesNode, signature.SignedAttributes)

	userDefinedAttributesNode := sigNode.Add("user defined attributes")
	addStringMapToTree(userDefinedAttributesNode, signature.UserDefinedAttributes)

	unsignedAttributesNode := sigNode.Add("unsigned attributes")
	for k, v := range signature.UnsignedAttributes {
		switch value := v.(type) {
		case string:
			unsignedAttributesNode.AddPair(k, value)
		case envelope.TimestampInfo:
			timestampNode := unsignedAttributesNode.Add("timestamp signature")
			if value.Error != "" {
				timestampNode.AddPair("error", value.Error)
				break
			}
			timestampNode.AddPair("timestamp", value.Timestamp)
			addCertificatesToTree(timestampNode, "certificates", value.Certificates)
		}
	}

	addCertificatesToTree(sigNode, "certificates", signature.Certificates)

	artifactNode := sigNode.Add("signed artifact")
	artifactNode.AddPair("media type", signature.SignedArtifact.MediaType)
	artifactNode.AddPair("digest", signature.SignedArtifact.Digest.String())
	artifactNode.AddPair("size", strconv.FormatInt(signature.SignedArtifact.Size, 10))

	sigNode.Print()
	return nil
}

func addMapToTree(node *tree.Node, m map[string]any) {
	if len(m) > 0 {
		for k, v := range m {
			node.AddPair(k, v)
		}
	} else {
		node.Add("(empty)")
	}
}

func addStringMapToTree(node *tree.Node, m map[string]string) {
	if len(m) > 0 {
		for k, v := range m {
			node.AddPair(k, v)
		}
	} else {
		node.Add("(empty)")
	}
}

func addCertificatesToTree(node *tree.Node, name string, certs []envelope.CertificateInfo) {
	certListNode := node.Add(name)
	for _, cert := range certs {
		certNode := certListNode.AddPair("SHA256 fingerprint", cert.SHA256Fingerprint)
		certNode.AddPair("issued to", cert.IssuedTo)
		certNode.AddPair("issued by", cert.IssuedBy)
		certNode.AddPair("expiry", cert.Expiry)
	}
}

func parseEnvelopeMediaType(filename string) (string, error) {
	parts := strings.Split(filename, ".")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid signature filename: %s", filename)
	}
	mediaType := strings.ToLower(parts[len(parts)-2])
	switch mediaType {
	case "jws":
		return jws.MediaTypeEnvelope, nil
	case "cose":
		return cose.MediaTypeEnvelope, nil
	default:
		return "", fmt.Errorf("unsupported signature format: %s", mediaType)
	}
}
