package main

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	"github.com/notaryproject/notation-go-lib"
	"github.com/notaryproject/notation-go-lib/crypto/cryptoutil"
	"github.com/notaryproject/notation-go-lib/signature/jws"
	"github.com/notaryproject/notation/pkg/cache"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/notaryproject/notation/pkg/crypto"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/urfave/cli/v2"
)

var verifyCommand = &cli.Command{
	Name:      "verify",
	Usage:     "Verifies OCI Artifacts",
	ArgsUsage: "<reference>",
	Flags: []cli.Flag{
		signatureFlag,
		&cli.StringSliceFlag{
			Name:    "cert",
			Aliases: []string{"c"},
			Usage:   "certificate names for verification",
		},
		&cli.StringSliceFlag{
			Name:      "cert-file",
			Usage:     "certificate files for verification",
			TakesFile: true,
		},
		&cli.StringSliceFlag{
			Name:  "ca-cert",
			Usage: "CA certificate names for verification",
		},
		&cli.StringSliceFlag{
			Name:      "ca-cert-file",
			Usage:     "CA certificate files for verification",
			TakesFile: true,
		},
		&cli.BoolFlag{
			Name:  "pull",
			Usage: "pull remote signatures before verification",
			Value: true,
		},
		localFlag,
		usernameFlag,
		passwordFlag,
		plainHTTPFlag,
		mediaTypeFlag,
	},
	Action: runVerify,
}

func runVerify(ctx *cli.Context) error {
	// initialize
	verifier, err := getVerifier(ctx)
	if err != nil {
		return err
	}
	manifestDesc, err := getManifestDescriptorFromContext(ctx)
	if err != nil {
		return err
	}

	sigPaths := ctx.StringSlice(signatureFlag.Name)
	if len(sigPaths) == 0 {
		if !ctx.Bool(localFlag.Name) && ctx.Bool("pull") {
			if err := pullSignatures(ctx, digest.Digest(manifestDesc.Digest)); err != nil {
				return err
			}
		}
		manifestDigest := digest.Digest(manifestDesc.Digest)
		sigDigests, err := cache.SignatureDigests(manifestDigest)
		if err != nil {
			return err
		}
		for _, sigDigest := range sigDigests {
			sigPaths = append(sigPaths, config.SignaturePath(manifestDigest, sigDigest))
		}
	}

	// core process
	if err := verifySignatures(ctx.Context, verifier, manifestDesc, sigPaths); err != nil {
		return err
	}

	// write out
	fmt.Println(manifestDesc.Digest)
	return nil
}

func verifySignatures(ctx context.Context, verifier notation.Verifier, manifestDesc ocispec.Descriptor, sigPaths []string) error {
	if len(sigPaths) == 0 {
		return errors.New("verification failure: no signatures found")
	}

	desc := convertDescriptorToNotation(manifestDesc)
	var opts notation.VerifyOptions
	var lastErr error
	for _, path := range sigPaths {
		sig, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		actualDesc, _, err := verifier.Verify(ctx, sig, opts)
		if err != nil {
			lastErr = fmt.Errorf("verification failure: %v", err)
			continue
		}

		if actualDesc != desc {
			lastErr = fmt.Errorf("verification failure: %s", desc.Digest)
			continue
		}
		return nil
	}
	return lastErr
}

func getVerifier(ctx *cli.Context) (notation.Verifier, error) {
	// resolve paths
	certPaths := ctx.StringSlice("cert-file")
	certPaths, err := appendCertPathFromName(certPaths, ctx.StringSlice("cert"))
	if err != nil {
		return nil, err
	}
	caCertPath := ctx.StringSlice("ca-cert-file")
	caCertPath, err = appendCertPathFromName(caCertPath, ctx.StringSlice("ca-cert"))
	if err != nil {
		return nil, err
	}
	if len(certPaths) == 0 && len(caCertPath) == 0 {
		cfg, err := config.LoadOrDefaultOnce()
		if err != nil {
			return nil, err
		}
		if len(cfg.VerificationCertificates.Certificates) == 0 {
			return nil, errors.New("trust certificate not specified")
		}
		for _, ref := range cfg.VerificationCertificates.Certificates {
			certPaths = append(certPaths, ref.Path)
		}
	}

	// read cert files
	var keys []*jws.VerificationKey
	roots := x509.NewCertPool()
	for _, path := range certPaths {
		bundledCerts, err := cryptoutil.ReadCertificateFile(path)
		if err != nil {
			return nil, err
		}
		for _, cert := range bundledCerts {
			keyID, err := crypto.KeyID(cert.PublicKey)
			if err != nil {
				return nil, err
			}
			key, err := jws.NewVerificationKey(cert.PublicKey, keyID)
			if err != nil {
				return nil, err
			}
			keys = append(keys, key)
			roots.AddCert(cert)
		}
	}
	for _, path := range caCertPath {
		bundledCerts, err := cryptoutil.ReadCertificateFile(path)
		if err != nil {
			return nil, err
		}
		for _, cert := range bundledCerts {
			roots.AddCert(cert)
		}
	}

	// construct verifier
	verifier := jws.NewVerifier(keys)
	verifier.VerifyOptions.Roots = roots
	return verifier, nil
}

func appendCertPathFromName(paths, names []string) ([]string, error) {
	for _, name := range names {
		cfg, err := config.LoadOrDefaultOnce()
		if err != nil {
			return nil, err
		}
		path, ok := cfg.VerificationCertificates.Certificates.Get(name)
		if !ok {
			return nil, errors.New("verification certificate not found: " + name)
		}
		paths = append(paths, path)
	}
	return paths, nil
}
