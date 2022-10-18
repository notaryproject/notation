package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/notaryproject/notation-core-go/testhelper"
	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation/internal/certificate"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/internal/slices"
	"github.com/notaryproject/notation/pkg/configutil"
)

func generateTestCert(opts *certGenerateTestOpts) error {
	// initialize
	host := opts.host
	name := opts.name
	if name == "" {
		name = host
	}

	// generate RSA private key
	bits := opts.bits
	fmt.Println("generating RSA Key with", bits, "bits")
	key, keyBytes, err := generateTestKey(bits)
	if err != nil {
		return err
	}

	rsaCertTuple, certBytes, err := generateSelfSignedCert(key, host)
	if err != nil {
		return err
	}
	fmt.Println("generated certificate expiring on", rsaCertTuple.Cert.NotAfter.Format(time.RFC3339))

	// write private key
	keyPath, certPath := dir.Path.Localkey(name)
	if err := osutil.WriteFileWithPermission(keyPath, keyBytes, 0600, false); err != nil {
		return fmt.Errorf("failed to write key file: %v", err)
	}
	fmt.Println("wrote key:", keyPath)

	// write the self-signed certificate
	if err := osutil.WriteFileWithPermission(certPath, certBytes, 0644, false); err != nil {
		return fmt.Errorf("failed to write certificate file: %v", err)
	}
	fmt.Println("wrote certificate:", certPath)

	// update config
	signingKeys, err := configutil.LoadSigningkeysOnce()
	if err != nil {
		return err
	}
	cfg, err := configutil.LoadConfigOnce()
	if err != nil {
		return err
	}
	isDefault := opts.isDefault
	keySuite := config.KeySuite{
		Name: name,
		X509KeyPair: &config.X509KeyPair{
			KeyPath:         keyPath,
			CertificatePath: certPath,
		},
	}
	err = addKeyCore(signingKeys, keySuite, isDefault)
	if err != nil {
		return err
	}
	trust := opts.trust
	// Add to the trust store
	if trust {
		if err := certificate.AddCertCore(certPath, "ca", name, true); err != nil {
			return err
		}
	}
	addCertToConfig(cfg, name, certPath)
	if err := cfg.Save(""); err != nil {
		return err
	}
	if err := signingKeys.Save(""); err != nil {
		return err
	}

	// write out
	fmt.Printf("%s: added to the key list\n", name)
	if isDefault {
		fmt.Printf("%s: marked as default\n", name)
	}
	return nil
}

func generateTestKey(bits int) (*rsa.PrivateKey, []byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})
	return key, keyPEM, nil
}

func generateCertPEM(rsaCertTuple *testhelper.RSACertTuple) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rsaCertTuple.Cert.Raw})
}

// generateTestCert generates a self-signed non-CA certificate
func generateSelfSignedCert(privateKey *rsa.PrivateKey, host string) (testhelper.RSACertTuple, []byte, error) {
	rsaCertTuple := testhelper.GetRSASelfSignedCertTupleWithPK(privateKey, host)
	return rsaCertTuple, generateCertPEM(&rsaCertTuple), nil
}

// addCertToConfig adds a self-signed certificate to config.json
// TODO: Notation will use certificates in the trust store to do verification.
// 		 Remove this path once trust policy is merged.
func addCertToConfig(cfg *config.Config, name, path string) {
	if !slices.Contains(cfg.VerificationCertificates.Certificates, name) {
		cfg.VerificationCertificates.Certificates = append(cfg.VerificationCertificates.Certificates, config.CertificateReference{
			Name: name,
			Path: path,
		})
	}
}
