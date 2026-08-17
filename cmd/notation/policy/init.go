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
	"github.com/notaryproject/notation/v2/cmd/notation/internal/display"
	"github.com/notaryproject/notation/v2/cmd/notation/internal/display/output"
	"github.com/notaryproject/notation/v2/internal/osutil"
	"github.com/spf13/cobra"
)

// wildcardRegistryScope is the registry scope that matches any registry. It is
// the default scope for a starter policy so verification applies everywhere
// until the user narrows it down.
const wildcardRegistryScope = "*"

type initOpts struct {
	printer           *output.Printer
	name              string
	registryScopes    []string
	trustStores       []string
	trustedIdentities []string
	force             bool
}

func initCmd() *cobra.Command {
	opts := initOpts{}
	command := &cobra.Command{
		Use:   `init [flags] --name <policy_name> --trust-store "<store_type>:<store_name>" --trusted-identity "<trusted_identity>"`,
		Short: "Initialize OCI trust policy configuration",
		Long: `Initialize OCI trust policy configuration.

The generated policy statement applies to all registries by default (registry scope "*"). Use --registry-scope to pin it to specific repositories.

Example - init an OCI trust policy configuration with a trust store and a trusted identity:
  notation policy init --name examplePolicy --trust-store ca:exampleStore --trusted-identity "x509.subject: C=US, ST=WA, O=acme-rockets.io"

Example - init an OCI trust policy configuration scoped to specific repositories:
  notation policy init --name examplePolicy --registry-scope registry.acme-rockets.io/software/net-monitor --trust-store ca:exampleStore --trusted-identity "x509.subject: C=US, ST=WA, O=acme-rockets.io"

Example - init an OCI trust policy configuration with multiple trust stores and trusted identities:
  notation policy init --name examplePolicy --trust-store ca:exampleStore --trust-store ca:exampleStore2 --trusted-identity "x509.subject: C=US, ST=WA, O=acme-rockets.io" --trusted-identity "x509.subject: C=US, ST=WA, L=Seattle, O=wabbit-networks.io"

Example - init an OCI trust policy configuration with any trusted identity:
  notation policy init --name examplePolicy --trust-store ca:exampleStore --trusted-identity "*"

Example - init an OCI trust policy configuration without prompt:
  notation policy init --name examplePolicy --trust-store ca:exampleStore --trusted-identity "x509.subject: C=US, ST=WA, O=acme-rockets.io" --force
`,
		Args: cobra.ExactArgs(0),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.printer = output.NewPrinter(cmd.OutOrStdout(), cmd.OutOrStderr())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(&opts)
		},
	}

	command.Flags().StringVarP(&opts.name, "name", "n", "", "name of the OCI trust policy")
	command.Flags().StringArrayVar(&opts.registryScopes, "registry-scope", []string{wildcardRegistryScope}, "registry scope the policy applies to, e.g. \"registry.acme-rockets.io/software/net-monitor\"; defaults to \"*\" for all registries")
	command.Flags().StringArrayVar(&opts.trustStores, "trust-store", nil, "trust store in the format \"<store_type>:<store_name>\"")
	command.Flags().StringArrayVar(&opts.trustedIdentities, "trusted-identity", nil, "trusted identity, use the format \"x509.subject:<subject_of_signing_certificate>\" for x509 CA scheme and \"<signing_authority_identity>\" for x509 signingAuthority scheme")
	command.Flags().BoolVar(&opts.force, "force", false, "override the existing OCI trust policy configuration, never prompt (default --force=false)")
	command.MarkFlagRequired("name")
	command.MarkFlagRequired("trust-store")
	command.MarkFlagRequired("trusted-identity")
	return command
}

func runInit(opts *initOpts) error {
	ociPolicy := trustpolicy.OCIDocument{
		Version: "1.0",
		TrustPolicies: []trustpolicy.OCITrustPolicy{
			{
				Name: opts.name,
				SignatureVerification: trustpolicy.SignatureVerification{
					VerificationLevel: trustpolicy.LevelStrict.Name,
				},
				RegistryScopes:    opts.registryScopes,
				TrustStores:       opts.trustStores,
				TrustedIdentities: opts.trustedIdentities,
			},
		},
	}
	if err := ociPolicy.Validate(); err != nil {
		return fmt.Errorf("invalid OCI policy: %w", err)
	}

	// optional confirmation
	if _, err := trustpolicy.LoadOCIDocument(); err == nil {
		if !opts.force {
			confirmed, err := display.AskForConfirmation(os.Stdin, "The OCI trust policy configuration already exists, do you want to overwrite it?", opts.force)
			if err != nil {
				return err
			}
			if !confirmed {
				return nil
			}
		} else {
			opts.printer.PrintErrorf("Warning: existing OCI trust policy configuration will be overwritten\n")
		}
	}

	policyPath, err := dir.ConfigFS().SysPath(dir.PathOCITrustPolicy)
	if err != nil {
		return fmt.Errorf("failed to obtain path of OCI trust policy configuration: %w", err)
	}
	policyJSON, err := json.MarshalIndent(ociPolicy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal OCI trust policy: %w", err)
	}
	if err = osutil.WriteFile(policyPath, policyJSON); err != nil {
		return fmt.Errorf("failed to write OCI trust policy configuration: %w", err)
	}

	return opts.printer.Printf("Successfully initialized OCI trust policy file to %s.\n", policyPath)
}
