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
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/notaryproject/notation-core-go/revocation/purpose"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/log"
	"github.com/notaryproject/notation/cmd/notation/internal/display"
	"github.com/notaryproject/notation/cmd/notation/internal/flag"
	"github.com/notaryproject/notation/cmd/notation/internal/sign"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/internal/httputil"
	"github.com/notaryproject/notation/internal/osutil"
	clirev "github.com/notaryproject/notation/internal/revocation"
	nx509 "github.com/notaryproject/notation/internal/x509"
	"github.com/notaryproject/tspclient-go"
	"github.com/spf13/cobra"
)

// timestampingTimeout is the timeout when requesting timestamp countersignature
// from a TSA
const timestampingTimeout = 15 * time.Second

type blobSignOpts struct {
	flag.LoggingFlagOpts
	flag.SignerFlagOpts
	expiry                 time.Duration
	pluginConfig           []string
	userMetadata           []string
	blobPath               string
	blobMediaType          string
	signatureDirectory     string
	tsaServerURL           string
	tsaRootCertificatePath string
	force                  bool
}

func signCommand(opts *blobSignOpts) *cobra.Command {
	if opts == nil {
		opts = &blobSignOpts{}
	}
	longMessage := `Produce a detached signature for a given blob.

The signature file will be written to the currently working directory with file name "{blob file name}.{signature format}.sig".

Note: a signing key must be specified. This can be done temporarily by specifying a key ID, or a new key can be configured using the command "notation key add"

Example - Sign a blob artifact using the default signing key, with the default JWS envelope, and store the signature at current directory:
  notation blob sign <blob_path>

Example - Sign a blob artifact by generating the signature in a particular directory: 
  notation blob sign --signature-directory <signature_directory_path> <blob_path>

Example - Sign a blob artifact and skip user confirmations when overwriting existing signature:
  notation blob sign --force <blob_path> 

Example - Sign a blob artifact using the default signing key, with the COSE envelope:
  notation blob sign --signature-format cose <blob_path>

Example - Sign a blob artifact with a specified plugin and signing key stored in KMS: 
  notation blob sign --plugin <plugin_name> --id <remote_key_id> <blob_path>

Example - Sign a blob artifact and add a user metadata to payload: 
  notation blob sign --user-metadata <metadata> <blob_path>

Example - Sign a blob artifact using a specified media type: 
  notation blob sign --media-type <media type> <blob_path>

Example - Sign a blob artifact using a specified key: 
  notation blob sign --key <key_name> <blob_path>

Example - Sign a blob artifact and specify the signature expiry duration, for example 24 hours: 
  notation blob sign --expiry 24h <blob_path>

Example - Sign a blob artifact with timestamping:
  notation blob sign --timestamp-url <TSA_url> --timestamp-root-cert <TSA_root_certificate_filepath> <blob_path>
`

	command := &cobra.Command{
		Use:   "sign [flags] <blob_path>",
		Short: "Produce a detached signature for a given blob",
		Long:  longMessage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing file path to the blob artifact: use `notation blob sign --help` to see what parameters are required")
			}
			opts.blobPath = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// timestamping
			if cmd.Flags().Changed("timestamp-url") {
				if opts.tsaServerURL == "" {
					return errors.New("timestamping: tsa url cannot be empty")
				}
				if opts.tsaRootCertificatePath == "" {
					return errors.New("timestamping: tsa root certificate path cannot be empty")
				}
			}
			return runBlobSign(cmd, opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	opts.SignerFlagOpts.ApplyFlagsToCommand(command)
	flag.SetPflagExpiry(command.Flags(), &opts.expiry)
	flag.SetPflagPluginConfig(command.Flags(), &opts.pluginConfig)
	flag.SetPflagUserMetadata(command.Flags(), &opts.userMetadata, flag.PflagUserMetadataSignUsage)
	command.Flags().StringVar(&opts.blobMediaType, "media-type", "application/octet-stream", "media type of the blob")
	command.Flags().StringVar(&opts.signatureDirectory, "signature-directory", ".", "directory where the blob signature needs to be placed")
	command.Flags().StringVar(&opts.tsaServerURL, "timestamp-url", "", "RFC 3161 Timestamping Authority (TSA) server URL")
	command.Flags().StringVar(&opts.tsaRootCertificatePath, "timestamp-root-cert", "", "filepath of timestamp authority root certificate")
	command.Flags().BoolVar(&opts.force, "force", false, "override the existing signature file, never prompt")
	command.MarkFlagsRequiredTogether("timestamp-url", "timestamp-root-cert")
	return command
}

func runBlobSign(command *cobra.Command, cmdOpts *blobSignOpts) error {
	// set log level
	ctx := cmdOpts.LoggingFlagOpts.InitializeLogger(command.Context())
	logger := log.GetLogger(ctx)

	blobSigner, err := sign.GetSigner(ctx, &cmdOpts.SignerFlagOpts)
	if err != nil {
		return err
	}
	blobOpts, err := prepareBlobSigningOpts(ctx, cmdOpts)
	if err != nil {
		return err
	}
	blobFile, err := os.Open(cmdOpts.blobPath)
	if err != nil {
		return err
	}
	defer blobFile.Close()

	// core process
	sig, _, err := notation.SignBlob(ctx, blobSigner, blobFile, blobOpts)
	if err != nil {
		return err
	}
	signaturePath := signatureFilepath(cmdOpts.signatureDirectory, cmdOpts.blobPath, cmdOpts.SignatureFormat)
	logger.Infof("Writing signature to file %s", signaturePath)

	// optional confirmation
	if !cmdOpts.force {
		if _, err := os.Stat(signaturePath); err == nil {
			confirmed, err := display.AskForConfirmation(os.Stdin, "The signature file already exists, do you want to overwrite it?", cmdOpts.force)
			if err != nil {
				return err
			}
			if !confirmed {
				return nil
			}
		}
	} else {
		fmt.Fprintln(os.Stderr, "Warning: existing signature file will be overwritten")
	}

	// write signature to file
	if err := osutil.WriteFile(signaturePath, sig); err != nil {
		return fmt.Errorf("failed to write signature to file: %w", err)
	}
	fmt.Printf("Successfully signed %s\n ", cmdOpts.blobPath)
	fmt.Printf("Signature file written to %s\n", signaturePath)
	return nil
}

func prepareBlobSigningOpts(ctx context.Context, opts *blobSignOpts) (notation.SignBlobOptions, error) {
	logger := log.GetLogger(ctx)

	mediaType, err := envelope.GetEnvelopeMediaType(opts.SignerFlagOpts.SignatureFormat)
	if err != nil {
		return notation.SignBlobOptions{}, err
	}
	pluginConfig, err := flag.ParseFlagMap(opts.pluginConfig, flag.PflagPluginConfig.Name)
	if err != nil {
		return notation.SignBlobOptions{}, err
	}
	userMetadata, err := flag.ParseFlagMap(opts.userMetadata, flag.PflagUserMetadata.Name)
	if err != nil {
		return notation.SignBlobOptions{}, err
	}
	signBlobOpts := notation.SignBlobOptions{
		SignerSignOptions: notation.SignerSignOptions{
			SignatureMediaType: mediaType,
			ExpiryDuration:     opts.expiry,
			PluginConfig:       pluginConfig,
		},
		ContentMediaType: opts.blobMediaType,
		UserMetadata:     userMetadata,
	}
	if opts.tsaServerURL != "" {
		// timestamping
		logger.Infof("Configured to timestamp with TSA %q", opts.tsaServerURL)
		signBlobOpts.Timestamper, err = tspclient.NewHTTPTimestamper(httputil.NewClient(ctx, &http.Client{Timeout: timestampingTimeout}), opts.tsaServerURL)
		if err != nil {
			return notation.SignBlobOptions{}, fmt.Errorf("cannot get http timestamper for timestamping: %w", err)
		}
		signBlobOpts.TSARootCAs, err = nx509.NewRootCertPool(opts.tsaRootCertificatePath)
		if err != nil {
			return notation.SignBlobOptions{}, err
		}
		tsaRevocationValidator, err := clirev.NewRevocationValidator(ctx, purpose.Timestamping)
		if err != nil {
			return notation.SignBlobOptions{}, fmt.Errorf("failed to create timestamping revocation validator: %w", err)
		}
		signBlobOpts.TSARevocationValidator = tsaRevocationValidator
	}
	return signBlobOpts, nil
}

// signatureFilepath returns the path to the signature file.
func signatureFilepath(signatureDirectory, blobPath, signatureFormat string) string {
	blobFilename := filepath.Base(blobPath)
	signatureFilename := fmt.Sprintf("%s.%s.sig", blobFilename, signatureFormat)
	return filepath.Join(signatureDirectory, signatureFilename)
}
