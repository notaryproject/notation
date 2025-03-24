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

package cert

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation/v2/cmd/notation/internal/display"
	"github.com/notaryproject/notation/v2/cmd/notation/internal/truststore"
	"github.com/spf13/cobra"
)

type certCleanupTestOpts struct {
	name      string
	confirmed bool
}

func certCleanupTestCommand(opts *certCleanupTestOpts) *cobra.Command {
	if opts == nil {
		opts = &certCleanupTestOpts{}
	}
	command := &cobra.Command{
		Use:   "cleanup-test [flags] <key_name>",
		Short: `Clean up a test RSA key and its corresponding certificate that were generated using the "generate-test" command.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing certificate common_name")
			}
			opts.name = args[0]
			return nil
		},
		Long: `Clean up a test RSA key and its corresponding certificate that were generated using the "generate-test" command.
Example - Clean up a test key and corresponding certificate named "wabbit-networks.io":
  notation cert cleanup-test "wabbit-networks.io"
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cleanupTestCert(opts)
		},
	}
	command.Flags().BoolVarP(&opts.confirmed, "yes", "y", false, "do not prompt for confirmation")
	return command
}

func cleanupTestCert(opts *certCleanupTestOpts) error {
	name := opts.name
	if !truststore.IsValidFileName(name) {
		return errors.New("name needs to follow [a-zA-Z0-9_.-]+ format")
	}
	prompt := fmt.Sprintf("Are you sure you want to clean up test key %q and its certificate?", name)
	confirmed, err := display.AskForConfirmation(os.Stdin, prompt, opts.confirmed)
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}

	// 1. remove from trust store
	localKeyPath, localCertPath := dir.LocalKeyPath(name)
	configFS := dir.ConfigFS()
	certPath, err := configFS.SysPath(localCertPath)
	if err != nil {
		return err
	}
	if err := truststore.DeleteCert("ca", name, filepath.Base(certPath), true); err != nil {
		return err
	}

	// 2. remove key and certificate files from LocalKeyPath
	keyPath, err := configFS.SysPath(localKeyPath)
	if err != nil {
		return err
	}

	if err := os.Remove(keyPath); err != nil {
		return err
	}
	if err := os.Remove(certPath); err != nil {
		return err
	}
	fmt.Printf("Successfully deleted %s and %s\n", filepath.Base(keyPath), filepath.Base(certPath))

	// 3. remove from signingkeys.json config
	exec := func(s *config.SigningKeys) error {
		_, err := s.Remove(name)
		return err
	}
	if err := config.LoadExecSaveSigningKeys(exec); err != nil {
		return err
	}
	fmt.Printf("Successfully removed %q from signingkeys.json\n", name)

	return nil
}
