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
	"github.com/notaryproject/notation/cmd/notation/internal/outputs"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/notaryproject/notation/internal/tree"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type blobInspectOpts struct {
	cmd.LoggingFlagOpts
	signaturePath string
	outputFormat  string
}

func inspectCommand(opts *blobInspectOpts) *cobra.Command {
	if opts == nil {
		opts = &blobInspectOpts{}
	}
	longMessage := `Inspect signature associated with the signed blob.

Example - Inspect BLOB signature:
  notation blob inspect <signature_path>

Example - Inspect BLOB signature and output as JSON:
  notation blob inspect --output json <signature_path>
`

	command := &cobra.Command{
		Use:   "blob inspect [signaturePath]",
		Short: "Inspect signature associated with the signed BLOB",
		Long:  longMessage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing signature path to the artifact: use `notation blob inspect --help` to see what parameters are required")
			}
			opts.signaturePath = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBlobInspect(opts)
		},
	}

	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	cmd.SetPflagOutput(command.Flags(), &opts.outputFormat, cmd.PflagOutputUsage)
	return command
}

func runBlobInspect(opts *blobInspectOpts) error {
	if opts.outputFormat != cmd.OutputJSON && opts.outputFormat != cmd.OutputPlaintext {
		return fmt.Errorf("unrecognized output format %s", opts.outputFormat)
	}

	// initialize
	mediaType, err := envelope.GetEnvelopeMediaType(filepath.Ext(opts.signaturePath))
	if err != nil {
		return err
	}
	contents, err := readFile(opts.signaturePath)
	if err != nil {
		return err
	}
	output := outputs.InspectOutput{MediaType: mediaType, Signatures: []outputs.SignatureOutput{}}
	skippedSignatures := false
	err, skippedSignatures, output.Signatures = outputs.Signature(mediaType, skippedSignatures, "nil", output, contents)
	if err != nil {
		return nil
	}
	if err := printOutput(opts.outputFormat, opts.signaturePath, output); err != nil {
		return err
	}
	return nil
}

func printOutput(outputFormat string, ref string, output outputs.InspectOutput) error {
	if outputFormat == cmd.OutputJSON {
		return ioutil.PrintObjectAsJSON(output)
	}

	if len(output.Signatures) == 0 {
		fmt.Printf("%s has no associated signature\n", ref)
		return nil
	}

	root := tree.New(ref)
	var signature outputs.SignatureOutput
	root.Add(signature.Digest)
	root.AddPair("media type", signature.MediaType)
	root.AddPair("signature algorithm", signature.SignatureAlgorithm)

	signedAttributesNode := root.Add("signed attributes")
	outputs.AddMapToTree(signedAttributesNode, signature.SignedAttributes)

	userDefinedAttributesNode := root.Add("user defined attributes")
	outputs.AddMapToTree(userDefinedAttributesNode, signature.UserDefinedAttributes)

	unsignedAttributesNode := root.Add("unsigned attributes")
	outputs.AddMapToTree(unsignedAttributesNode, signature.UnsignedAttributes)

	certListNode := root.Add("certificates")
	for _, cert := range signature.Certificates {
		certNode := certListNode.AddPair("SHA256 fingerprint", cert.SHA256Fingerprint)
		certNode.AddPair("issued to", cert.IssuedTo)
		certNode.AddPair("issued by", cert.IssuedBy)
		certNode.AddPair("expiry", cert.Expiry)
	}

	artifactNode := root.Add("signed artifact")
	artifactNode.AddPair("media type", signature.SignedArtifact.MediaType)
	artifactNode.AddPair("digest", signature.SignedArtifact.Digest.String())
	artifactNode.AddPair("size", strconv.FormatInt(signature.SignedArtifact.Size, 10))

	root.Print()
	return nil
}

func readFile(path string) ([]byte, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	size, err := os.Stat(path)
	if size.Size() == 0 {
		return nil, fmt.Errorf("file is empty")
	}
	r := strings.NewReader(string(file))
	var n int64 = 10485760 //10MB in bytes
	limitedReader := &io.LimitedReader{R: r, N: n}
	contents, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}
	if limitedReader.N == 0 {
		return nil, fmt.Errorf("unable to read as file size was greater than %v bytes", n)
	}
	return contents, nil
}
