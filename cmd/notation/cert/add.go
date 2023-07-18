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

	"github.com/notaryproject/notation/cmd/notation/internal/truststore"
	"github.com/spf13/cobra"
)

type certAddOpts struct {
	storeType  string
	namedStore string
	path       []string
}

func certAddCommand(opts *certAddOpts) *cobra.Command {
	if opts == nil {
		opts = &certAddOpts{}
	}
	command := &cobra.Command{
		Use:   "add --type <type> --store <name> [flags] <cert_path>...",
		Short: "Add certificates to the trust store.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing certificate path")
			}
			opts.path = args
			return nil
		},
		Long: `Add certificates to the trust store

Example - Add a certificate to the "ca" type of a named store "acme-rockets":
  notation cert add --type ca --store acme-rockets acme-rockets.crt

Example - Add a certificate to the "signingAuthority" type of a named store "wabbit-networks":
  notation cert add --type signingAuthority --store wabbit-networks wabbit-networks.pem
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return addCerts(opts)
		},
	}
	command.Flags().StringVarP(&opts.storeType, "type", "t", "", "specify trust store type, options: ca, signingAuthority")
	command.Flags().StringVarP(&opts.namedStore, "store", "s", "", "specify named store")
	command.MarkFlagRequired("type")
	command.MarkFlagRequired("store")
	return command
}

func addCerts(opts *certAddOpts) error {
	storeType := opts.storeType
	if storeType == "" {
		return errors.New("store type cannot be empty")
	}
	if !truststore.IsValidStoreType(storeType) {
		return fmt.Errorf("unsupported store type: %s", storeType)
	}
	namedStore := opts.namedStore
	if !truststore.IsValidFileName(namedStore) {
		return errors.New("named store name needs to follow [a-zA-Z0-9_.-]+ format")
	}
	var success []string
	var failure []string
	var errorSlice []error
	for _, p := range opts.path {
		err := truststore.AddCert(p, storeType, namedStore, false)
		if err != nil {
			failure = append(failure, p)
			errorSlice = append(errorSlice, err)
		} else {
			success = append(success, p)
		}
	}

	//write out
	if len(success) != 0 {
		fmt.Printf("Successfully added following certificates to named store %s of type %s:\n", namedStore, storeType)
		for _, p := range success {
			fmt.Println(p)
		}
	}
	if len(failure) != 0 {
		errStr := fmt.Sprintf("Failed to add following certificates to named store %s of type %s:\n", namedStore, storeType)

		for ind := range failure {
			errStr = errStr + fmt.Sprintf("%s, with error %q\n", failure[ind], errorSlice[ind])
		}
		return errors.New(errStr)
	}

	return nil
}
