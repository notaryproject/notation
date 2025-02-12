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
	"strings"

	"github.com/notaryproject/notation-core-go/signature/cose"
	"github.com/notaryproject/notation-core-go/signature/jws"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/cmd/notation/internal/display"
	"github.com/notaryproject/notation/cmd/notation/internal/option"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/spf13/cobra"
)

type blobVerifyOpts struct {
	cmd.LoggingFlagOpts
	option.Common
	blobPath            string
	signaturePath       string
	pluginConfig        []string
	userMetadata        []string
	policyStatementName string
	blobMediaType       string
}

func verifyCommand(opts *blobVerifyOpts) *cobra.Command {
	if opts == nil {
		opts = &blobVerifyOpts{}
	}
	longMessage := `Verify a signature associated with a blob.

Prerequisite: added a certificate into trust store and created a trust policy.

Example - Verify a signature on a blob artifact:
  notation blob verify --signature <signature_path> <blob_path>

Example - Verify the signature on a blob artifact with user metadata:
  notation blob verify --user-metadata <metadata> --signature <signature_path> <blob_path>

Example - Verify the signature on a blob artifact with media type:
  notation blob verify --media-type <media_type> --signature <signature_path> <blob_path>
 
Example - Verify the signature on a blob artifact using a policy statement name:
  notation blob verify --policy-name <policy_name> --signature <signature_path> <blob_path>
`
	command := &cobra.Command{
		Use:   "verify [flags] --signature <signature_path> <blob_path>",
		Short: "Verify a signature associated with a blob",
		Long:  longMessage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing path to the blob artifact: use `notation blob verify --help` to see what parameters are required")
			}
			opts.blobPath = args[0]
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.signaturePath == "" {
				return errors.New("filepath of the signature cannot be empty")
			}
			if cmd.Flags().Changed("media-type") && opts.blobMediaType == "" {
				return errors.New("--media-type is set but with empty value")
			}
			opts.Common.Parse(cmd)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVerify(cmd, opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	command.Flags().StringVar(&opts.signaturePath, "signature", "", "filepath of the signature to be verified")
	command.Flags().StringArrayVar(&opts.pluginConfig, "plugin-config", nil, "{key}={value} pairs that are passed as it is to a plugin, if the verification is associated with a verification plugin, refer plugin documentation to set appropriate values")
	command.Flags().StringVar(&opts.blobMediaType, "media-type", "", "media type of the blob to verify")
	command.Flags().StringVar(&opts.policyStatementName, "policy-name", "", "policy name to verify against. If not provided, the global policy is used if exists")
	cmd.SetPflagUserMetadata(command.Flags(), &opts.userMetadata, cmd.PflagUserMetadataVerifyUsage)
	command.MarkFlagRequired("signature")
	return command
}

func runVerify(command *cobra.Command, cmdOpts *blobVerifyOpts) error {
	// set log level
	ctx := cmdOpts.LoggingFlagOpts.InitializeLogger(command.Context())

	// initialize
	displayHandler := display.NewBlobVerifyHandler(cmdOpts.Printer)
	blobFile, err := os.Open(cmdOpts.blobPath)
	if err != nil {
		return err
	}
	defer blobFile.Close()

	signatureBytes, err := os.ReadFile(cmdOpts.signaturePath)
	if err != nil {
		return err
	}
	blobVerifier, err := cmd.GetVerifier(ctx, true)
	if err != nil {
		return err
	}

	// set up verification plugin config
	pluginConfigs, err := cmd.ParseFlagMap(cmdOpts.pluginConfig, cmd.PflagPluginConfig.Name)
	if err != nil {
		return err
	}

	// set up user metadata
	userMetadata, err := cmd.ParseFlagMap(cmdOpts.userMetadata, cmd.PflagUserMetadata.Name)
	if err != nil {
		return err
	}
	signatureMediaType, err := parseSignatureMediaType(cmdOpts.signaturePath)
	if err != nil {
		return err
	}
	verifyBlobOpts := notation.VerifyBlobOptions{
		BlobVerifierVerifyOptions: notation.BlobVerifierVerifyOptions{
			SignatureMediaType: signatureMediaType,
			PluginConfig:       pluginConfigs,
			UserMetadata:       userMetadata,
			TrustPolicyName:    cmdOpts.policyStatementName,
		},
		ContentMediaType: cmdOpts.blobMediaType,
	}
	_, outcome, err := notation.VerifyBlob(ctx, blobVerifier, blobFile, signatureBytes, verifyBlobOpts)
	outcomes := []*notation.VerificationOutcome{outcome}
	err = ioutil.PrintVerificationFailure(outcomes, cmdOpts.blobPath, err, true)
	if err != nil {
		return err
	}
	displayHandler.OnVerifySucceeded(outcomes, cmdOpts.blobPath)
	return displayHandler.Render()
}

// parseSignatureMediaType returns the media type of the signature file.
// `application/jose+json` and `application/cose` are supported.
func parseSignatureMediaType(signaturePath string) (string, error) {
	signatureFileName := filepath.Base(signaturePath)
	sigFilenameArr := strings.Split(signatureFileName, ".")

	// a valid signature file name has at least 3 parts.
	// for example, `myFile.jws.sig`
	if len(sigFilenameArr) < 3 {
		return "", fmt.Errorf("invalid signature filename %s", signatureFileName)
	}
	format := sigFilenameArr[len(sigFilenameArr)-2]
	switch format {
	case "cose":
		return cose.MediaTypeEnvelope, nil
	case "jws":
		return jws.MediaTypeEnvelope, nil
	}
	return "", fmt.Errorf("unsupported signature format %s", format)
}
