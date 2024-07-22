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

package x509

import (
	"testing"

	corex509 "github.com/notaryproject/notation-core-go/x509"
)

func TestIsRootCertificate(t *testing.T) {
	tsaRoot, err := corex509.ReadCertificateFile("../testdata/tsaRootCA.cer")
	if err != nil {
		t.Fatal(err)
	}
	isRoot, err := IsRootCertificate(tsaRoot[0])
	if err != nil {
		t.Fatal(err)
	}
	if !isRoot {
		t.Fatal("expected IsRootCertificate to return true")
	}

	intermediate, err := corex509.ReadCertificateFile("../testdata/intermediate.pem")
	if err != nil {
		t.Fatal(err)
	}
	expectedErrMsg := "crypto/rsa: verification error"
	_, err = IsRootCertificate(intermediate[0])
	if err == nil || err.Error() != expectedErrMsg {
		t.Fatalf("expected %s, but got %s", expectedErrMsg, err)
	}

	selfSigned, err := corex509.ReadCertificateFile("../testdata/self-signed.crt")
	if err != nil {
		t.Fatal(err)
	}
	expectedErrMsg = "x509: invalid signature: parent certificate cannot sign this kind of certificate"
	_, err = IsRootCertificate(selfSigned[0])
	if err == nil || err.Error() != expectedErrMsg {
		t.Fatalf("expected %s, but got %s", expectedErrMsg, err)
	}
}
