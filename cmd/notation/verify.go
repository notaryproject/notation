package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/slices"
	"github.com/notaryproject/notation/pkg/cache"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/notaryproject/notation/pkg/signature"
	"github.com/opencontainers/go-digest"
	"github.com/urfave/cli/v2"
)

var verifyCommand = &cli.Command{
	Name:      "verify",
	Usage:     "Verifies OCI Artifacts",
	ArgsUsage: "<reference>",
	Flags: []cli.Flag{
		flagSignature,
		&cli.StringSliceFlag{
			Name:    "cert",
			Aliases: []string{"c"},
			Usage:   "certificate names for verification",
		},
		&cli.StringSliceFlag{
			Name:      cmd.FlagCertFile.Name,
			Usage:     "certificate files for verification",
			TakesFile: true,
		},
		&cli.BoolFlag{
			Name:  "pull",
			Usage: "pull remote signatures before verification",
			Value: true,
		},
		flagLocal,
		flagUsername,
		flagPassword,
		flagPlainHTTP,
		flagMediaType,
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

	sigPaths := ctx.StringSlice(flagSignature.Name)
	if len(sigPaths) == 0 {
		if !ctx.Bool(flagLocal.Name) && ctx.Bool("pull") {
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

func verifySignatures(ctx context.Context, verifier notation.Verifier, manifestDesc notation.Descriptor, sigPaths []string) error {
	if len(sigPaths) == 0 {
		return errors.New("verification failure: no signatures found")
	}

	var opts notation.VerifyOptions
	var lastErr error
	for _, path := range sigPaths {
		sig, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		desc, err := verifier.Verify(ctx, sig, opts)
		if err != nil {
			lastErr = fmt.Errorf("verification failure: %v", err)
			continue
		}

		if !desc.Equal(manifestDesc) {
			lastErr = fmt.Errorf("verification failure: %s", manifestDesc.Digest)
			continue
		}
		return nil
	}
	return lastErr
}

func getVerifier(ctx *cli.Context) (notation.Verifier, error) {
	certPaths := ctx.StringSlice(cmd.FlagCertFile.Name)
	certPaths, err := appendCertPathFromName(certPaths, ctx.StringSlice("cert"))
	if err != nil {
		return nil, err
	}
	if len(certPaths) == 0 {
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
	return signature.NewVerifierFromFiles(certPaths)
}

func appendCertPathFromName(paths, names []string) ([]string, error) {
	for _, name := range names {
		cfg, err := config.LoadOrDefaultOnce()
		if err != nil {
			return nil, err
		}
		idx := slices.Index(cfg.VerificationCertificates.Certificates, name)
		if idx < 0 {
			return nil, errors.New("verification certificate not found: " + name)
		}
		paths = append(paths, cfg.VerificationCertificates.Certificates[idx].Path)
	}
	return paths, nil
}
