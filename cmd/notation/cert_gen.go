package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/urfave/cli/v2"
)

func generateTestCert(ctx *cli.Context) error {
	// initialize
	hosts := ctx.Args().Slice()
	if len(hosts) == 0 {
		return errors.New("missing certificate hosts")
	}
	name := ctx.String("name")
	if name == "" {
		name = hosts[0]
	}

	// generate RSA private key
	bits := ctx.Int("bits")
	fmt.Println("generating RSA Key with", bits, "bits")
	key, keyBytes, err := generateTestKey(bits)
	if err != nil {
		return err
	}

	// generate self-created certificate chain
	rootCA, rootBytes, rootPrivKey, err := generateTestRootCert(hosts, ctx.Duration("expiry"), bits)
	if err != nil {
		return err
	}
	leafCA, leafBytes, err := generateTestLeafCert(rootCA, rootPrivKey, &key.PublicKey, hosts, ctx.Duration("expiry"))
	if err != nil {
		return err
	}
	fmt.Println("generated certificates expiring on", leafCA.NotAfter.Format(time.RFC3339))

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
	isDefault := ctx.Bool(keyDefaultFlag.Name)
	keySuite := config.KeySuite{
		Name: name,
		X509KeyPair: &config.X509KeyPair{
			KeyPath:         keyPath,
			CertificatePath: certPath,
		},
	}
	err = addKeyCore(cfg, keySuite, ctx.Bool(keyDefaultFlag.Name))
	if err != nil {
		return err
	}
	trust := ctx.Bool("trust")
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

func generateCert(template, parent *x509.Certificate, publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey) (*x509.Certificate, []byte, error) {
	certBytes, err := x509.CreateCertificate(rand.Reader, template, parent, publicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %v", err)
	}
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("generated invalid certificate: %v", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	return cert, certPEM, nil
}

// generateTestRootCert generates a self-signed root certificate
func generateTestRootCert(hosts []string, expiry time.Duration, bits int) (*x509.Certificate, []byte, *rsa.PrivateKey, error) {
	now := time.Now()
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate root serial number: %v", err)
	}
	rootTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: hosts[0] + "RootCA",
		},
		NotBefore:             now,
		NotAfter:              now.Add(expiry),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate root key: %v", err)
	}
	rootCert, rootPEM, err := generateCert(&rootTemplate, &rootTemplate, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate root cert: %v", err)
	}
	return rootCert, rootPEM, priv, nil
}

// generateTestLeafCert generates the leaf certificate with rootCA as parent
func generateTestLeafCert(rootCA *x509.Certificate, rootKey *rsa.PrivateKey, publicKey *rsa.PublicKey, hosts []string, expiry time.Duration) (*x509.Certificate, []byte, error) {
	now := time.Now()
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate leaf serial number: %v", err)
	}
	leafTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: hosts[0] + "LeafCA",
		},
		NotBefore:             now,
		NotAfter:              now.Add(expiry),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	leafCert, leafPEM, err := generateCert(&leafTemplate, rootCA, publicKey, rootKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate leaf cert: %v", err)
	}
	return leafCert, leafPEM, nil
}
