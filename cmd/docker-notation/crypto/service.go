package crypto

import (
	"crypto/x509"

	"github.com/docker/libtrust"
	"github.com/notaryproject/notation-go-lib"
	x509n "github.com/notaryproject/notation-go-lib/signature/x509"
	"github.com/notaryproject/notation-go-lib/simple"
)

// GetSigningService returns a signing service
func GetSigningService(keyPath string, certPaths ...string) (notation.SigningService, error) {
	var (
		key         libtrust.PrivateKey
		commonCerts []*x509.Certificate
		rootCerts   *x509.CertPool
		err         error
	)
	if keyPath != "" {
		key, err = x509n.ReadPrivateKeyFile(keyPath)
		if err != nil {
			return nil, err
		}
	}
	if len(certPaths) != 0 {
		rootCerts = x509.NewCertPool()
		for _, certPath := range certPaths {
			certs, err := x509n.ReadCertificateFile(certPath)
			if err != nil {
				return nil, err
			}
			commonCerts = append(commonCerts, certs...)
			for _, cert := range certs {
				rootCerts.AddCert(cert)
			}
		}
	}
	return simple.NewSigningService(key, commonCerts, commonCerts, rootCerts)
}
