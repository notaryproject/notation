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
	"errors"
	"fmt"
	"github.com/notaryproject/notation/cmd/notation/internal/experimental"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/spf13/cobra"
)

func blobSignCommand(opts *signOpts) *cobra.Command {
	if opts == nil {
		opts = &signOpts{
			inputType: inputTypeRegistry, // remote registry by default
		}
	}
	longMessage := `Sign artifacts

Note: a signing key must be specified. This can be done temporarily by specifying a key ID, or a new key can be configured using the command "notation key add"

Example - Sign a BLOB artifact using the default signing key, with the default JWS envelope, and use BLOB image manifest to store the signature:
  notation blob sign <registry>/<repository>@<digest>

Example - Sign a BLOB artifact using the default signing key, with the COSE envelope:
  notation blob sign --signature-format cose <registry>/<repository>@<digest> 

Example - Sign a BLOB artifact with a specified plugin and signing key stored in KMS 
  notation blob sign --plugin <plugin_name> --id <remote_key_id> <registry>/<repository>@<digest>

Example - Sign a BLOB artifact using a specified key
  notation blob sign --key <key_name> <registry>/<repository>@<digest>

Example - Sign a BLOB artifact identified by a tag (Notation will resolve tag to digest)
  notation blob sign <registry>/<repository>:<tag>

Example - Sign a BLOB artifact stored in a registry and specify the signature expiry duration, for example 24 hours
  notation blob sign --expiry 24h <registry>/<repository>@<digest>
`

	command := &cobra.Command{
		Use:   "blob sign [flags] <reference>",
		Short: "Sign artifacts",
		Long:  longMessage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing reference")
			}
			opts.reference = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSign(cmd, opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SignerFlagOpts.ApplyFlagsToCommand(command)
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	cmd.SetPflagExpiry(command.Flags(), &opts.expiry)
	cmd.SetPflagPluginConfig(command.Flags(), &opts.pluginConfig)
	cmd.SetPflagUserMetadata(command.Flags(), &opts.userMetadata, cmd.PflagUserMetadataSignUsage)
	cmd.SetPflagReferrersAPI(command.Flags(), &opts.allowReferrersAPI, fmt.Sprintf(cmd.PflagReferrersUsageFormat, "sign"))
	return command
}

func blobInspectCommand(opts *inspectOpts) *cobra.Command {
	if opts == nil {
		opts = &inspectOpts{}
	}
	longMessage := `Inspect all signatures associated with the signed artifact.

Example - Inspect signatures on an BLOB artifact identified by a digest:
  notation blob inspect <registry>/<repository>@<digest>

Example - Inspect signatures on an BLOB artifact identified by a tag  (Notation will resolve tag to digest):
  notation blob inspect <registry>/<repository>:<tag>

Example - Inspect signatures on an BLOB artifact identified by a digest and output as json:
  notation blob inspect --output json <registry>/<repository>@<digest>
`
	experimentalExamples := `
Example - [Experimental] Inspect signatures on an BLOB artifact identified by a digest using the Referrers API, if not supported (returns 404), fallback to the Referrers tag schema
  notation blob inspect --allow-referrers-api <registry>/<repository>@<digest>
`
	command := &cobra.Command{
		Use:   "blob inspect [reference]",
		Short: "Inspect all signatures associated with the signed artifact",
		Long:  longMessage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing reference")
			}
			opts.reference = args[0]
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return experimental.CheckFlagsAndWarn(cmd, "allow-referrers-api")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.maxSignatures <= 0 {
				return fmt.Errorf("max-signatures value %d must be a positive number", opts.maxSignatures)
			}
			return runInspect(cmd, opts)
		},
	}

	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	cmd.SetPflagOutput(command.Flags(), &opts.outputFormat, cmd.PflagOutputUsage)
	command.Flags().IntVar(&opts.maxSignatures, "max-signatures", 100, "maximum number of signatures to evaluate or examine")
	cmd.SetPflagReferrersAPI(command.Flags(), &opts.allowReferrersAPI, fmt.Sprintf(cmd.PflagReferrersUsageFormat, "inspect"))
	experimental.HideFlags(command, experimentalExamples, []string{"allow-referrers-api"})
	return command
}
