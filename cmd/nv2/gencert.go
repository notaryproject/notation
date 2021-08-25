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
	"log"
	"math/big"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

const (
	// This needs to be configurable. Once the location of the
	// configuration is finalized this parameter and file locations
	// should be exposed as options on the cli with mock tests.
	KEYS_BASE_DIR = ".notation"
)

var certificatesCommand = &cli.Command{
	Name:  "certificates",
	Usage: "Commands to manage certificates",
	Subcommands: []*cli.Command{
		generateCertCommand,
	},
}

var generateCertCommand = &cli.Command{
	Name:      "generate",
	Usage:     "Generates a test crt and key file.",
	ArgsUsage: "[host]",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:     "rsaBits",
			Usage:    "--rsaBits 3072",
			Required: false,
			Value:    3072,
		},
		&cli.StringFlag{
			Name:     "not-after",
			Usage:    "--not-after 2006-01-02T15:04:05-07:00 (default is 1 year)",
			Required: false,
		},
	},

	Action: runGenerateCert,
}

func ensureKeysDir() (string, error) {

	// Expected to ensure ~/.notation/keys
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	keysDir := filepath.Join(dirname, KEYS_BASE_DIR, "keys")
	fsStat, err := os.Stat(keysDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(keysDir, 0700) // Only user has permissions on this directory
		if err != nil {
			return "", err
		}
	} else if fsStat.IsDir() == false {
		return "", fmt.Errorf("%s should be a directory", keysDir)
	}

	return keysDir, nil
}

//ref: https://golang.org/src/crypto/tls/generate_cert.go

func runGenerateCert(ctx *cli.Context) error {

	host := ctx.Args().First()
	if len(host) == 0 {
		return errors.New("Missing required [host] parameter")
	}

	// Set certificate validity
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Duration(365 * 24 * time.Hour))

	expiry := ctx.String("not-after")
	var err error
	if len(expiry) != 0 {
		notAfter, err = time.Parse(time.RFC3339, expiry)
		if err != nil {
			return fmt.Errorf("Invalid --not-after %s value specified %v", expiry, err)
		}
	}

	if notAfter.Before(notBefore) {
		return fmt.Errorf("Invalid --not-after that is earlier than not-before [%s] specified", notBefore.Format(time.RFC3339))
	}

	// Generate RSA Bits
	rsaBits := ctx.Int("rsaBits")
	var priv crypto.Signer

	fmt.Printf("Generating RSA Key with %d bits\n", rsaBits)
	priv, err = rsa.GenerateKey(rand.Reader, rsaBits)

	if err != nil {
		return fmt.Errorf("Failed to generate private key: %v", err)
	}

	// ECDSA, ED25519 and RSA subject keys should have the DigitalSignature
	// KeyUsage bits set in the x509.Certificate template
	keyUsage := x509.KeyUsageDigitalSignature
	extKeyUsage := []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("Failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           extKeyUsage,
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	template.Subject = pkix.Name{
		Organization: []string{hosts[0]},
		CommonName:   hosts[0],
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, priv.Public(), priv)
	if err != nil {
		return fmt.Errorf("Failed to create certificate: %v", err)
	}

	// Write the crt public key file
	keysDir, err := ensureKeysDir()
	if err != nil {
		return fmt.Errorf("Could not access keys directory: %v", err)
	}

	crtFileName := path.Join(keysDir, hosts[0]+".crt")
	crtFilePath, err := filepath.Abs(crtFileName)
	if err != nil {
		return fmt.Errorf("Unable to get full path of the file: %v", err)
	}

	certOut, err := os.OpenFile(crtFileName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("Failed to open %s for writing: %v", crtFilePath, err)
	}

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("Failed to write data to %s: %v", crtFilePath, err)
	}
	if err := certOut.Close(); err != nil {
		return fmt.Errorf("Error closing %s: %v", crtFilePath, err)
	}

	fmt.Printf("Generated certificates expiring on %s\n", notAfter.Format(time.RFC3339))
	fmt.Printf("Wrote self-signed certificate file: %s\n", crtFilePath)

	// Write the private key file
	keyFileName := path.Join(keysDir, hosts[0]+".key")
	keyFilePath, err := filepath.Abs(keyFileName)
	if err != nil {
		return fmt.Errorf("Unable to get full path of the file: %v", err)
	}

	keyOut, err := os.OpenFile(keyFilePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("Failed to open key.pem for writing: %v", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return fmt.Errorf("Unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("Failed to write data to key.pem: %v", err)
	}
	if err := keyOut.Close(); err != nil {
		return fmt.Errorf("Error closing %s: %v", keyFilePath, err)
	}
	fmt.Printf("Wrote private key file: %s\n", keyFilePath)
	return nil
}
