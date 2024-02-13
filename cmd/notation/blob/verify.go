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
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/verifier"
	"github.com/notaryproject/notation/cmd/notation/internal/sharedutils"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/spf13/cobra"
)

type blobVerifyOpts struct {
	cmd.LoggingFlagOpts
	signaturePath        string
	blobPath             string
	pluginConfig         []string
	userMetadata         []string
	maxSignatureAttempts int
	policyName           string
}

func verifyCommand(opts *blobVerifyOpts) *cobra.Command {
	if opts == nil {
		opts = &blobVerifyOpts{}
	}
	longMessage := `Verify BLOB artifacts

Prerequisite: added a certificate into trust store and created a trust policy.

Example - Verify a signature on a BLOB artifact:
  notation blob verify --signature <signaturePath> <blobPath>

Example - Verify the signature on a BLOB artifact with user metadata:
  notation blob verify --user-metadata <metadata> --signature <signaturePath> <blobPath>

Example - Verify the signature on a BLOB artifact with media type:
  notation blob verify --media-type <media_type> --signature <signaturePath> <blobPath>
 
Example - Verify the signature on a BLOB artifact using a policy name:
  notation blob verify --policy-name <policy_name> --signature <signaturePath> <blobPath>
`
	command := &cobra.Command{
		Use:   "blob verify [flags] --signature <signaturePath> <blobPath>",
		Short: "Verify BLOB artifacts",
		Long:  longMessage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing signature/blob path to the artifact: use `notation blob verify --help` to see what parameters are required")
			}
			opts.signaturePath = args[0]
			opts.blobPath = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.maxSignatureAttempts <= 0 {
				return fmt.Errorf("max-signatures value %d must be a positive number", opts.maxSignatureAttempts)
			}
			return runVerify(cmd, opts)
		},
	}
	command.Flags().StringArrayVar(&opts.pluginConfig, "plugin-config", nil, "{key}={value} pairs that are passed as it is to a plugin, if the verification is associated with a verification plugin, refer plugin documentation to set appropriate values")
	cmd.SetPflagUserMetadata(command.Flags(), &opts.userMetadata, cmd.PflagUserMetadataVerifyUsage)
	command.Flags().IntVar(&opts.maxSignatureAttempts, "max-signatures", 100, "maximum number of signatures to evaluate or examine")
	return command
}

func runVerify(command *cobra.Command, opts *blobVerifyOpts) error {
	// set log level
	ctx := opts.LoggingFlagOpts.InitializeLogger(command.Context())

	// initialize
	sigVerifier, err := verifier.NewFromConfig()
	if err != nil {
		return err
	}

	// set up verification plugin config.
	configs, err := cmd.ParseFlagMap(opts.pluginConfig, cmd.PflagPluginConfig.Name)
	if err != nil {
		return err
	}

	// set up user metadata
	userMetadata, err := cmd.ParseFlagMap(opts.userMetadata, cmd.PflagUserMetadata.Name)
	if err != nil {
		return err
	}

	// Todo: we will need to replace signer with actual blob signer implementation in notation-go
	// core verify process
	verifyOpts := notation.VerifyOptions{
		PluginConfig:         configs,
		MaxSignatureAttempts: opts.maxSignatureAttempts,
		UserMetadata:         userMetadata,
	}
	_, outcomes, err := notation.BlobVerify(ctx, sigVerifier, verifyOpts) //PlaceHolder
	printOut := "placeholder"
	err = sharedutils.CheckVerificationFailure(outcomes, printOut, err)
	if err != nil {
		return err
	}
	sharedutils.ReportVerificationSuccess(outcomes, printOut)
	return nil
}
