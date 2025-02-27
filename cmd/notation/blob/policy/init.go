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

package policy

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/notaryproject/notation/cmd/notation/internal/cmdutil"
	"github.com/notaryproject/notation/cmd/notation/internal/option"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/spf13/cobra"
)

type initOpts struct {
	option.Common
	name            string
	trustStore      string
	trustedIdentity string
	force           bool
}

func initCmd() *cobra.Command {
	opts := initOpts{}
	command := &cobra.Command{
		Use:   "init [flags]",
		Short: "Init blob trust policy file",
		Long: `Init blob trust policy file.

Example - init a blob trust file with trust store and trust policy:
  notation blob policy init --trust-store <store-type>:<store-name> --trusted-policy file "x509.subject: C=US, ST=WA, L=Seattle, O=acme-rockets.io, OU=Finance, CN=SecureBuilder"
`,
		Args: cobra.ExactArgs(0),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.Common.Parse(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(&opts)
		},
	}

	command.Flags().StringVarP(&opts.name, "name", "n", "", "name of the blob trust policy")
	command.Flags().StringVarP(&opts.trustStore, "trust-store", "s", "", "trust store in format <store-type>:<store-name>")
	command.Flags().StringVarP(&opts.trustedIdentity, "trusted-identity", "i", "", "trusted identity (e.g. \"x509.subject: C=US, ST=WA, L=Seattle, O=acme-rockets.io\")")
	command.Flags().BoolVar(&opts.force, "force", false, "override the existing blob trust policy configuration without prompt")
	command.MarkFlagRequired("name")
	command.MarkFlagRequired("trust-store")
	command.MarkFlagRequired("trusted-identity")
	return command
}

func runInit(opts *initOpts) error {
	blobPolicy := trustpolicy.BlobDocument{
		Version: "1.0",
		TrustPolicies: []trustpolicy.BlobTrustPolicy{
			{
				Name: opts.name,
				SignatureVerification: trustpolicy.SignatureVerification{
					VerificationLevel: "strict",
				},
				TrustStores:       []string{opts.trustStore},
				TrustedIdentities: []string{opts.trustedIdentity},
				GlobalPolicy:      true,
			},
		},
	}

	if err := blobPolicy.Validate(); err != nil {
		return fmt.Errorf("invalid blob policy: %w", err)
	}
	policyJSON, err := json.MarshalIndent(blobPolicy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal blob trust policy: %w", err)
	}

	// optional confirmation
	if _, err = trustpolicy.LoadBlobDocument(); err == nil {
		if !opts.force {
			confirmed, err := cmdutil.AskForConfirmation(os.Stdin, "The blob trust policy configuration already exists, do you want to overwrite it?", opts.force)
			if err != nil {
				return err
			}
			if !confirmed {
				return nil
			}
		} else {
			opts.Printer.PrintErrorf("Warning: existing blob trust policy configuration will be overwritten")
		}
	}

	policyPath, err := dir.ConfigFS().SysPath(dir.PathBlobTrustPolicy)
	if err != nil {
		return fmt.Errorf("failed to obtain path of blob trust policy configuration: %w", err)
	}
	if err = osutil.WriteFile(policyPath, policyJSON); err != nil {
		return fmt.Errorf("failed to write blob trust policy configuration: %w", err)
	}

	return opts.Printer.Printf("Successfully initialized blob trust policy file to %s.\n", policyPath)
}
