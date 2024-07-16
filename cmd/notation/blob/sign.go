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
	"context"
	"errors"
	"fmt"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

type blobSignOpts struct {
	cmd.LoggingFlagOpts
	cmd.SignerFlagOpts
	expiry             time.Duration
	pluginConfig       []string
	userMetadata       []string
	blobPath           string
	signatureDirectory string
	force              bool
}

func signCommand(opts *blobSignOpts) *cobra.Command {
	if opts == nil {
		opts = &blobSignOpts{}
	}
	longMessage := `Sign BLOB artifacts

Note: a signing key must be specified. This can be done temporarily by specifying a key ID, or a new key can be configured using the command "notation key add"

Example - Sign a BLOB artifact using the default signing key, with the default JWS envelope, and use BLOB image manifest to store the signature:
  notation blob sign <blob_path>

Example - Sign a BLOB artifact by generating the signature in a particular directory: 
  notation blob sign --signature-directory <directory_path> <blob_path>

Example - Sign a BLOB artifact and skip user confirmations when overwriting existing signature:
  notation blob sign --force <blob_path> 

Example - Sign a BLOB artifact using the default signing key, with the COSE envelope:
  notation blob sign --signature-format cose <blob_path>

Example - Sign a BLOB artifact with a specified plugin and signing key stored in KMS: 
  notation blob sign --plugin <plugin_name> --id <remote_key_id> <blob_path>

Example - Sign a BLOB artifact and add a user metadata to payload: 
  notation blob sign --user-metadata <metadata> <blob_path>

Example - Sign a BLOB artifact using a specified media type: 
  notation blob sign --media-type <media type> <blob_path>

Example - Sign a BLOB artifact using a specified key: 
  notation blob sign --key <key_name> <blob_path>

Example - Sign a BLOB artifact and specify the signature expiry duration, for example 24 hours: 
  notation blob sign --expiry 24h <blob_path>
`

	command := &cobra.Command{
		Use:   "blob sign [flags] <blobPath>",
		Short: "Sign BLOB artifact",
		Long:  longMessage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing blob file path to the artifact: use `notation blob sign --help` to see what parameters are required")
			}
			opts.blobPath = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBlobSign(cmd, opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SignerFlagOpts.ApplyFlagsToCommand(command)
	cmd.SetPflagExpiry(command.Flags(), &opts.expiry)
	cmd.SetPflagPluginConfig(command.Flags(), &opts.pluginConfig)
	cmd.SetPflagUserMetadata(command.Flags(), &opts.userMetadata, cmd.PflagUserMetadataSignUsage)
	return command
}

func runBlobSign(command *cobra.Command, cmdOpts *blobSignOpts) error {
	// set log level
	ctx := cmdOpts.LoggingFlagOpts.InitializeLogger(command.Context())

	// Todo: we will need to replace signer with actual blob signer implementation in notation-go
	// initialize
	signer, err := cmd.GetSigner(ctx, &cmdOpts.SignerFlagOpts)
	if err != nil {
		return err
	}
	blobOpts, err := prepareBlobSigningOpts(ctx, cmdOpts)
	if err != nil {
		return err
	}
	contents, err := os.ReadFile(cmdOpts.blobPath)
	if err != nil {
		return err
	}
	// core process
	_, _, err = notation.SignBlob(ctx, signer, strings.NewReader(string(contents)), blobOpts) //PlaceHolder
	if err != nil {
		return err
	}
	fmt.Println("Successfully signed")
	return nil
}

func prepareBlobSigningOpts(ctx context.Context, opts *blobSignOpts) (notation.SignOptions, error) {
	mediaType, err := envelope.GetEnvelopeMediaType(opts.SignerFlagOpts.SignatureFormat)
	if err != nil {
		return notation.SignOptions{}, err
	}
	pluginConfig, err := cmd.ParseFlagMap(opts.pluginConfig, cmd.PflagPluginConfig.Name)
	if err != nil {
		return notation.SignOptions{}, err
	}
	userMetadata, err := cmd.ParseFlagMap(opts.userMetadata, cmd.PflagUserMetadata.Name)
	if err != nil {
		return notation.SignOptions{}, err
	}
	blobOpts := notation.SignOptions{
		SignerSignOptions: notation.SignerSignOptions{
			SignatureMediaType: mediaType,
			ExpiryDuration:     opts.expiry,
			PluginConfig:       pluginConfig,
		},
		UserMetadata: userMetadata,
	}
	return blobOpts, nil
}
