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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
)

// Show shows trust policy configuration.
//
// - If isOciPolicy is true, it will show OCI trust policy configuration.
// Otherwise, it will show blob trust policy configuration.
func Show(isOCIPolicy bool) error {
	var (
		policyJSON []byte
		err        error
		doc        policy
	)
	if isOCIPolicy {
		doc = &trustpolicy.OCIDocument{}
		policyJSON, err = loadOCIDocument()
	} else {
		doc = &trustpolicy.BlobDocument{}
		policyJSON, err = loadBlobDocument()
	}
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("failed to show trust policy as the trust policy file does not exist.\nYou can import one using `notation policy import <path-to-policy.json>`")
		}
		return fmt.Errorf("failed to show trust policy: %w", err)
	}

	if err = json.Unmarshal(policyJSON, &doc); err == nil {
		err = doc.Validate()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "Existing trust policy configuration is invalid, you may update or create a new one via `notation policy import <path-to-policy.json>`\n")
		// not returning to show the invalid policy configuration
	}

	// show policy content
	_, err = os.Stdout.Write(policyJSON)
	return err
}

func loadOCIDocument() ([]byte, error) {
	f, err := dir.ConfigFS().Open(dir.PathOCITrustPolicy)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		f, err = dir.ConfigFS().Open(dir.PathTrustPolicy)
		if err != nil {
			return nil, err
		}
	}
	defer f.Close()
	return io.ReadAll(f)
}

func loadBlobDocument() ([]byte, error) {
	f, err := dir.ConfigFS().Open(dir.PathBlobTrustPolicy)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}
