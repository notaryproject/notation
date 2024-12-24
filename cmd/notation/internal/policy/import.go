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
	"github.com/notaryproject/notation/internal/osutil"
)

type policy interface {
	Validate() error
}

// Import imports trust policy configuration from a JSON file.
//
// - If force is true, it will override the existing trust policy configuration
// without prompting.
// - If isOciPolicy is true, it will import OCI trust policy configuration.
// Otherwise, it will import blob trust policy configuration.
func Import(filePath string, force, isOCIPolicy bool) error {
	// read configuration
	policyJSON, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read trust policy file: %w", err)
	}

	// parse and validate
	var doc policy = &trustpolicy.OCIDocument{}
	if !isOCIPolicy {
		doc = &trustpolicy.BlobDocument{}
	}
	if err = json.Unmarshal(policyJSON, doc); err != nil {
		return fmt.Errorf("failed to parse trust policy configuration: %w", err)
	}
	if err = doc.Validate(); err != nil {
		return fmt.Errorf("failed to validate trust policy: %w", err)
	}

	// optional confirmation
	if !force {
		if isOCIPolicy {
			_, err = trustpolicy.LoadDocument()
		} else {
			_, err = trustpolicy.LoadBlobDocument()
		}
		if err == nil {
			confirmed, err := cmdutil.AskForConfirmation(os.Stdin, "The trust policy file already exists, do you want to overwrite it?", force)
			if err != nil {
				return err
			}
			if !confirmed {
				return nil
			}
		}
	} else {
		fmt.Fprintln(os.Stderr, "Warning: existing trust policy configuration file will be overwritten")
	}

	// write
	trustPolicyName := dir.PathOCITrustPolicy
	if !isOCIPolicy {
		trustPolicyName = dir.PathBlobTrustPolicy
	}
	policyPath, err := dir.ConfigFS().SysPath(trustPolicyName)
	if err != nil {
		return fmt.Errorf("failed to obtain path of trust policy file: %w", err)
	}
	if err = osutil.WriteFile(policyPath, policyJSON); err != nil {
		return fmt.Errorf("failed to write trust policy file: %w", err)
	}

	// clear old trust policy
	if isOCIPolicy {
		oldPolicyPath, err := dir.ConfigFS().SysPath(dir.PathTrustPolicy)
		if err != nil {
			return fmt.Errorf("failed to obtain path of trust policy file: %w", err)
		}
		if err := osutil.RemoveIfExists(oldPolicyPath); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to clear old trust policy %q: %v\n", oldPolicyPath, err)
		}
	}

	_, err = fmt.Fprintln(os.Stdout, "Trust policy configuration imported successfully.")
	return err
}
