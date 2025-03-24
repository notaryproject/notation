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
	"io/fs"
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
				return errors.New("missing key name")
			}
			if args[0] == "" {
				return errors.New("key name cannot be empty")
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
	relativeKeyPath, relativeCertPath := dir.LocalKeyPath(name)
	configFS := dir.ConfigFS()
	certPath, _ := configFS.SysPath(relativeCertPath) // err is always nil
	certFileName := filepath.Base(certPath)
	err = truststore.DeleteCert("ca", name, certFileName, true)
	if err != nil {
		var pathError *fs.PathError
		if errors.As(err, &pathError) && errors.Is(pathError, fs.ErrNotExist) {
			fmt.Printf("Certificate %s does not exist in trust store %s of type ca.", certFileName, name)
		} else {
			return err
		}
	}

	// 2. remove key from signingkeys.json config
	exec := func(s *config.SigningKeys) error {
		_, err := s.Remove(name)
		return err
	}
	err = config.LoadExecSaveSigningKeys(exec)
	if err != nil {
		if errors.Is(err, config.KeyNotFoundError{KeyName: name}) {
			fmt.Printf("Key %s does not exist in signingkeys.json.", name)
		} else {
			return err
		}
	}
	fmt.Printf("Successfully removed key %s from signingkeys.json\n", name)

	// 3. remove key and certificate files from LocalKeyPath
	keyPath, _ := configFS.SysPath(relativeKeyPath) // err is always nil
	err = os.Remove(keyPath)
	if err != nil {
		var pathError *fs.PathError
		if errors.As(err, &pathError) && errors.Is(pathError, fs.ErrNotExist) {
			fmt.Printf("The key file %s does not exist.", keyPath)
		} else {
			return err
		}
	}
	err = os.Remove(certPath)
	if err != nil {
		var pathError *fs.PathError
		if errors.As(err, &pathError) && errors.Is(pathError, fs.ErrNotExist) {
			fmt.Printf("The certificate file %s does not exist.", certPath)
		} else {
			return err
		}
	}
	fmt.Println("Cleanup completed successfully.")
	return nil
}
