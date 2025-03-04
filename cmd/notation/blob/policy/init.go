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
	name              string
	trustStores       []string
	trustedIdentities []string
	force             bool
	global            bool
}

func initCmd() *cobra.Command {
	opts := initOpts{}
	command := &cobra.Command{
		Use:   `init [flags] --name <policy_name> --trust-store "<store_type>:<store_name>" --trusted-identity "<trusted_identity>"`,
		Short: "Initialize blob trust policy configuration",
		Long: `Initialize blob trust policy configuration.

Example - init a blob trust policy configuration with a trust store and a trusted identity:
  notation blob policy init --name examplePolicy --trust-store ca:exampleStore --trusted-identity "x509.subject: C=US, ST=WA, O=acme-rockets.io"
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
	command.Flags().StringArrayVar(&opts.trustStores, "trust-store", nil, "trust store in the format \"<store_type>:<store_name>\"")
	command.Flags().StringArrayVar(&opts.trustedIdentities, "trusted-identity", nil, "trusted identity, use the format \"x509.subject:<subject_of_signing_certificate>\" for x509 CA scheme and \"<signing_authority_identity>\" for x509 signingAuthority scheme")
	command.Flags().BoolVar(&opts.force, "force", false, "override the existing blob trust policy configuration, never prompt (default --force=false)")
	command.Flags().BoolVar(&opts.global, "global", false, "set the policy as the global policy (default --global=false)")
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
					VerificationLevel: trustpolicy.LevelStrict.Name,
				},
				TrustStores:       opts.trustStores,
				TrustedIdentities: opts.trustedIdentities,
				GlobalPolicy:      opts.global,
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
			opts.Printer.PrintErrorf("Warning: existing blob trust policy configuration will be overwritten\n")
		}
	}

	policyPath, err := dir.ConfigFS().SysPath(dir.PathBlobTrustPolicy)
	if err != nil {
		return fmt.Errorf("failed to obtain path of blob trust policy configuration: %w", err)
	}
	if err = osutil.WriteFile(policyPath, policyJSON); err != nil {
		return fmt.Errorf("failed to write blob trust policy configuration: %w", err)
	}

	return opts.Printer.Printf("Successfully initialized blob trust policy file to %s\n", policyPath)
}
