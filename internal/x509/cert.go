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
	"bytes"
	"crypto/x509"
)

// IsRootCertificate returns true if cert is a root certificate.
// A root certificate MUST be a self-signed and self-issued CA certificate with
// valid BasicConstraints.
func IsRootCertificate(cert *x509.Certificate) (bool, error) {
	// CheckSignatureFrom also checks cert.BasicConstraintsValid
	if err := cert.CheckSignatureFrom(cert); err != nil {
		return false, err
	}
	return cert.IsCA && bytes.Equal(cert.RawSubject, cert.RawIssuer), nil
}
