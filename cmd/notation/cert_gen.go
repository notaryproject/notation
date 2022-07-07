package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/spf13/cobra"
)

func generateTestCert(command *cobra.Command) error {
	// initialize
	hosts := command.Flags().Args()
	if len(hosts) == 0 {
		return errors.New("missing certificate hosts")
	}
	name, _ := command.Flags().GetString("name")
	if name == "" {
		name = hosts[0]
	}

	// generate RSA private key
	bits, _ := command.Flags().GetInt("bits")
	fmt.Println("generating RSA Key with", bits, "bits")
	key, keyBytes, err := generateTestKey(bits)
	if err != nil {
		return err
	}

	// generate self-signed certificate
	expiry, _ := command.Flags().GetDuration("expiry")
	cert, certBytes, err := generateTestSelfSignedCert(key, hosts, expiry)
	if err != nil {
		return err
	}
	fmt.Println("generated certificates expiring on", cert.NotAfter.Format(time.RFC3339))

	// write private key
	keyPath := config.KeyPath(name)
	if err := osutil.WriteFileWithPermission(keyPath, keyBytes, 0600, false); err != nil {
		return fmt.Errorf("failed to write key file: %v", err)
	}
	fmt.Println("wrote key:", keyPath)

	// write self-signed certificate
	certPath := config.CertificatePath(name)
	if err := osutil.WriteFileWithPermission(certPath, certBytes, 0644, false); err != nil {
		return fmt.Errorf("failed to write certificate file: %v", err)
	}
	fmt.Println("wrote certificate:", certPath)

	// update config
	cfg, err := config.LoadOrDefault()
	if err != nil {
		return err
	}
	isDefault, _ := command.Flags().GetBool(keyDefaultFlag.Name)
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
	trust, _ := command.Flags().GetBool("trust")
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

func generateTestKey(bits int) (crypto.Signer, []byte, error) {
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

func generateTestSelfSignedCert(key crypto.Signer, hosts []string, expiry time.Duration) (*x509.Certificate, []byte, error) {
	now := time.Now()
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: hosts[0],
		},
		NotBefore:             now,
		NotAfter:              now.Add(expiry),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
		BasicConstraintsValid: true,
	}
	for _, host := range hosts {
		if ip := net.ParseIP(host); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, host)
		}
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, key.Public(), key)
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
