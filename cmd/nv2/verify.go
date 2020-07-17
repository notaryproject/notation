package main

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"

	"github.com/notaryproject/nv2/internal/crypto"
	"github.com/notaryproject/nv2/pkg/signature"
	x509nv2 "github.com/notaryproject/nv2/pkg/signature/x509"
	"github.com/urfave/cli/v2"
)

var verifyCommand = &cli.Command{
	Name:      "verify",
	Usage:     "verifies artifacts or images",
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
			Name:      "ca-cert",
			Aliases:   []string{"c"},
			Usage:     "CA certs for verification",
			TakesFile: true,
		},
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
	var roots *x509.CertPool
	if caCerts := ctx.StringSlice("ca-cert"); len(caCerts) > 0 {
		roots = x509.NewCertPool()
		for _, path := range caCerts {
			certs, err := crypto.ReadCertificateFile(path)
			if err != nil {
				return nil, err
			}
			for _, cert := range certs {
				roots.AddCert(cert)
			}
		}
	}

	verifier, err := x509nv2.NewVerifier(roots)
	if err != nil {
		return nil, err
	}

	scheme := signature.NewScheme()
	scheme.RegisterVerifier(verifier)
	return scheme, nil
}
