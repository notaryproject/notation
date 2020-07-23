package main

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"

	"github.com/notaryproject/nv2/internal/crypto"
	"github.com/notaryproject/nv2/pkg/signature"
	"github.com/notaryproject/nv2/pkg/signature/gpg"
	x509nv2 "github.com/notaryproject/nv2/pkg/signature/x509"
	"github.com/urfave/cli/v2"
)

var verifyCommand = &cli.Command{
	Name:      "verify",
	Usage:     "verifies OCI Artifacts",
	ArgsUsage: "[<scheme://reference>]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:      "signature",
			Aliases:   []string{"s", "f"},
			Usage:     "signature file",
			Required:  true,
			TakesFile: true,
		},
		&cli.StringSliceFlag{
			Name:      "cert",
			Aliases:   []string{"c"},
			Usage:     "certs for verification [x509]",
			TakesFile: true,
		},
		&cli.StringSliceFlag{
			Name:      "ca-cert",
			Usage:     "CA certs for verification [x509]",
			TakesFile: true,
		},
		&cli.StringFlag{
			Name:      "key-ring",
			Usage:     "gpg public key ring file [gpg]",
			Value:     gpg.DefaultPublicKeyRingPath(),
			TakesFile: true,
		},
		&cli.BoolFlag{
			Name:  "disable-gpg",
			Usage: "disable GPG for verification [gpg]",
		},
		usernameFlag,
		passwordFlag,
		insecureFlag,
	},
	Action: runVerify,
}

func runVerify(ctx *cli.Context) error {
	// initialize
	scheme, err := getSchemeForVerification(ctx)
	if err != nil {
		return err
	}
	sigma, err := readSignatrueFile(ctx.String("signature"))
	if err != nil {
		return err
	}

	// core process
	content, _, err := scheme.Verify(sigma)
	if err != nil {
		return fmt.Errorf("verification failure: %v", err)
	}
	manifest, err := getManifestFromContext(ctx)
	if err != nil {
		return err
	}
	if !containsManifest(content.Manifests, manifest) {
		return fmt.Errorf("verification failure: manifest is not signed: %s", manifest.Digest)
	}

	// write out
	fmt.Println(manifest.Digest)
	return nil
}

func containsManifest(set []signature.Manifest, target signature.Manifest) bool {
	for _, manifest := range set {
		if manifest.Digest == target.Digest && manifest.Size == target.Size {
			return true
		}
	}
	return false
}

func readSignatrueFile(path string) (sig signature.Signed, err error) {
	file, err := os.Open(path)
	if err != nil {
		return sig, err
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&sig)
	return sig, err
}

func getSchemeForVerification(ctx *cli.Context) (*signature.Scheme, error) {
	scheme := signature.NewScheme()

	// add x509 verifier
	verifier, err := getX509Verifier(ctx)
	if err != nil {
		return nil, err
	}
	scheme.RegisterVerifier(verifier)

	// add gpg verifier
	if !ctx.Bool("disable-gpg") {
		verifier, err := gpg.NewVerifier(ctx.String("key-ring"))
		if err != nil {
			return nil, err
		}
		scheme.RegisterVerifier(verifier)
	}

	return scheme, nil
}

func getX509Verifier(ctx *cli.Context) (signature.Verifier, error) {
	roots := x509.NewCertPool()

	var certs []*x509.Certificate
	for _, path := range ctx.StringSlice("cert") {
		bundledCerts, err := crypto.ReadCertificateFile(path)
		if err != nil {
			return nil, err
		}
		certs = append(certs, bundledCerts...)
		for _, cert := range bundledCerts {
			roots.AddCert(cert)
		}
	}
	for _, path := range ctx.StringSlice("ca-cert") {
		bundledCerts, err := crypto.ReadCertificateFile(path)
		if err != nil {
			return nil, err
		}
		for _, cert := range bundledCerts {
			roots.AddCert(cert)
		}
	}

	return x509nv2.NewVerifier(certs, roots)
}
