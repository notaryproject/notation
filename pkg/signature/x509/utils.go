package x509

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"

	"github.com/docker/libtrust"
)

// ReadPrivateKeyFile reads a key PEM file as a libtrust key
func ReadPrivateKeyFile(path string) (libtrust.PrivateKey, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("no PEM data found")
	}
	switch block.Type {
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return libtrust.FromCryptoPrivateKey(key)
	default:
		return libtrust.UnmarshalPrivateKeyPEM(raw)
	}
}

// ReadCertificateFile reads a certificate PEM file
func ReadCertificateFile(path string) ([]*x509.Certificate, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var certs []*x509.Certificate
	block, rest := pem.Decode(raw)
	for block != nil {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
		block, rest = pem.Decode(rest)
	}
	return certs, nil
}
