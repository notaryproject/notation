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
	"path/filepath"
	"strings"

	"github.com/notaryproject/notation-core-go/signature/cose"
	"github.com/notaryproject/notation-core-go/signature/jws"
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
			signatureMediaType, err := parseSignatureMediaType(signaturePath)
			if err != nil {
				Fail(err.Error())
			}
			notation.Exec("blob", "verify", "--signature", signatureMediaType, blobPath).
				MatchKeyWords(VerifySuccessfully)
		})
	})
})

func parseSignatureMediaType(signaturePath string) (string, error) {
	signatureFileName := filepath.Base(signaturePath)
	format := strings.Split(signatureFileName, ".")[1]
	switch format {
	case "cose":
		return cose.MediaTypeEnvelope, nil
	case "jws":
		return jws.MediaTypeEnvelope, nil
	}
	return "", fmt.Errorf("unsupported signature format %s", format)
}

func signatureFilepath(signatureDirectory, blobPath, signatureFormat string) string {
	blobFilename := filepath.Base(blobPath)
	signatureFilename := fmt.Sprintf("%s.%s.sig", blobFilename, signatureFormat)
	return filepath.Join(signatureDirectory, signatureFilename)
}
