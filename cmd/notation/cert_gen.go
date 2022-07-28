package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/notaryproject/notation-core-go/testhelper"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/pkg/config"
)

func generateTestCert(opts *certGenerateTestOpts) error {
	// initialize
	hosts := opts.hosts
	name := opts.name
	if name == "" {
		name = hosts[0]
	}

	// generate RSA private key
	bits := opts.bits
	fmt.Println("generating RSA Key with", bits, "bits")
	key, keyBytes, err := generateTestKey(bits)
	if err != nil {
		return err
	}

	// generate self-created certificate chain
	rsaRootCertTuple, rootBytes, err := generateTestRootCert(hosts, bits)
	if err != nil {
		return err
	}
	rsaLeafCertTuple, leafBytes, err := generateTestLeafCert(&rsaRootCertTuple, key, hosts)
	if err != nil {
		return err
	}
	fmt.Println("generated certificates expiring on", rsaLeafCertTuple.Cert.NotAfter.Format(time.RFC3339))

	// write private key
	keyPath := config.KeyPath(name)
	if err := osutil.WriteFileWithPermission(keyPath, keyBytes, 0600, false); err != nil {
		return fmt.Errorf("failed to write key file: %v", err)
	}
	fmt.Println("wrote key:", keyPath)

	// write self-signed certificate
	certPath := config.CertificatePath(name)
	if err := osutil.WriteFileWithPermission(certPath, append(leafBytes, rootBytes...), 0644, false); err != nil {
		return fmt.Errorf("failed to write certificate file: %v", err)
	}
	fmt.Println("wrote certificate:", certPath)

	// update config
	cfg, err := config.LoadOrDefault()
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
	err = addKeyCore(cfg, keySuite, isDefault)
	if err != nil {
		return err
	}
	trust := opts.trust
	if trust {
		if err := addCertCore(cfg, name, certPath); err != nil {
			return err
		}
	}
	if err := cfg.Save(); err != nil {
		return err
	}

	// write out
	fmt.Printf("%s: added to the key list\n", name)
	if isDefault {
		fmt.Printf("%s: marked as default\n", name)
	}
	if trust {
		fmt.Printf("%s: added to the certificate list\n", name)
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

// generateTestRootCert generates a self-signed root certificate
func generateTestRootCert(hosts []string, bits int) (testhelper.RSACertTuple, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return testhelper.RSACertTuple{}, nil, fmt.Errorf("failed to generate root key: %v", err)
	}
	rsaRootCertTuple := testhelper.GetRSACertTupleWithPK(priv, hosts[0]+"RootCA", nil)
	return rsaRootCertTuple, generateCertPEM(&rsaRootCertTuple), nil
}

// generateTestLeafCert generates the leaf certificate
func generateTestLeafCert(rsaRootCertTuple *testhelper.RSACertTuple, privateKey *rsa.PrivateKey, hosts []string) (testhelper.RSACertTuple, []byte, error) {
	rsaLeafCertTuple := testhelper.GetRSACertTupleWithPK(privateKey, hosts[0]+"LeafCA", rsaRootCertTuple)
	return rsaLeafCertTuple, generateCertPEM(&rsaLeafCertTuple), nil
}
