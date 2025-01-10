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
	"fmt"
	"os"
	"path/filepath"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation blob verify", func() {
	// Success cases
	It("with blob verify", func() {
		HostWithBlob(BaseBlobOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			blobDir := filepath.Dir(blobPath)
			notation.Exec("blob", "sign", "--force", "--signature-directory", blobDir, blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")

			signaturePath := signatureFilepath(blobDir, blobPath, "jws")
			notation.Exec("blob", "verify", "--signature", signaturePath, blobPath).
				MatchKeyWords(VerifySuccessfully)
		})
	})

	// Failure cases
	It("with blob verify no permission to read blob", func() {
		HostWithBlob(BaseBlobOptions(), func(notation *utils.ExecOpts, blobPath string, vhost *utils.VirtualHost) {
			noPermissionBlobPath := filepath.Join(vhost.AbsolutePath(), "noPermissionBlob")
			newBlobFile, err := os.Create(noPermissionBlobPath)
			if err != nil {
				Fail(err.Error())
			}
			defer newBlobFile.Close()

			blobDir := filepath.Dir(noPermissionBlobPath)
			notation.Exec("blob", "sign", "--force", "--signature-directory", blobDir, blobPath).
				MatchKeyWords(SignSuccessfully).
				MatchKeyWords("Signature file written to")
			if err := os.Chmod(noPermissionBlobPath, 0000); err != nil {
				Fail(err.Error())
			}
			defer os.Chmod(noPermissionBlobPath, 0700)

			signaturePath := signatureFilepath(blobDir, blobPath, "jws")
			notation.ExpectFailure().Exec("blob", "verify", "--signature", signaturePath, blobPath).
				MatchErrKeyWords("permission denied")
		})
	})
})

func signatureFilepath(signatureDirectory, blobPath, signatureFormat string) string {
	blobFilename := filepath.Base(blobPath)
	signatureFilename := fmt.Sprintf("%s.%s.sig", blobFilename, signatureFormat)
	return filepath.Join(signatureDirectory, signatureFilename)
}
