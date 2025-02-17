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

	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/spf13/cobra"
)

type initOpts struct {
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
  notation blob policy init --trust-store <store-type>:<store-name> --trusted policy file "x509.subject: C=US, ST=WA, L=Seattle, O=acme-rockets.io, OU=Finance, CN=SecureBuilder"
`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(opts)
		},
	}

	command.Flags().StringVar(&opts.trustStore, "trust-store", "", "trust store in format <store-type>:<store-name>")
	command.Flags().StringVar(&opts.trustedIdentity, "trusted-identity", "", "trust identity (e.g. \"x509.subject: C=US, ST=WA, L=Seattle, O=acme-rockets.io\")")
	command.Flags().BoolVar(&opts.force, "force", false, "override the existing blob trust policy configuration without prompt")
	command.MarkFlagRequired("trust-store")
	command.MarkFlagRequired("trusted-identity")

	return command
}

func runInit(opts initOpts) error {
	blobPolicy := trustpolicy.BlobDocument{
		Version: "1.0",
		TrustPolicies: []trustpolicy.BlobTrustPolicy{
			{
				Name: "default-policy",
				SignatureVerification: trustpolicy.SignatureVerification{
					VerificationLevel: "strict",
				},
				TrustStores:       []string{opts.trustStore},
				TrustedIdentities: []string{opts.trustedIdentity},
				GlobalPolicy:      true,
			},
		},
	}

	// Validate the policy
	if err := blobPolicy.Validate(); err != nil {
		return fmt.Errorf("invalid blob policy: %w", err)
	}

	policyJson, err := json.MarshalIndent(blobPolicy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal blob trust policy: %w", err)
	}

	if err := writeBlobTrustPolicy(policyJson, opts.force); err != nil {
		return err
	}

	_, err = fmt.Fprintln(os.Stdout, "Successfully initialized blob trust policy file.")
	return err
}
