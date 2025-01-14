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
	"fmt"

	corex509 "github.com/notaryproject/notation-core-go/x509"
)

// IsRootCertificate returns true if cert is a root certificate.
// A root certificate MUST be a self-signed and self-issued certificate.
func IsRootCertificate(cert *x509.Certificate) (bool, error) {
	if err := cert.CheckSignatureFrom(cert); err != nil {
		return false, err
	}
	return bytes.Equal(cert.RawSubject, cert.RawIssuer), nil
}

// NewRootCertPool returns a new x509 CertPool containing the root certificate
// from rootCertificatePath.
func NewRootCertPool(rootCertificatePath string) (*x509.CertPool, error) {
	rootCerts, err := corex509.ReadCertificateFile(rootCertificatePath)
	if err != nil {
		return nil, err
	}
	if len(rootCerts) == 0 {
		return nil, fmt.Errorf("cannot find any certificate from %q. Expecting single x509 root certificate in PEM or DER format from the file", rootCertificatePath)
	}
	if len(rootCerts) > 1 {
		return nil, fmt.Errorf("found more than one certificates from %q. Expecting single x509 root certificate in PEM or DER format from the file", rootCertificatePath)
	}
	rootCert := rootCerts[0]
	isRoot, err := IsRootCertificate(rootCert)
	if err != nil {
		return nil, fmt.Errorf("failed to check root certificate with error: %w", err)
	}
	if !isRoot {
		return nil, fmt.Errorf("certificate from %q is not a root certificate. Expecting single x509 root certificate in PEM or DER format from the file", rootCertificatePath)
	}
	rootCAs := x509.NewCertPool()
	rootCAs.AddCert(rootCert)
	return rootCAs, nil
}
