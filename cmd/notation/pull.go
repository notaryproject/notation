package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/pkg/cache"
	"github.com/notaryproject/notation/pkg/config"
	notationregistry "github.com/notaryproject/notation/pkg/registry"
	"github.com/opencontainers/go-digest"
	"github.com/urfave/cli/v2"
	"oras.land/oras-go/v2/registry"
)

var pullCommand = &cli.Command{
	Name:      "pull",
	Usage:     "Pull signatures from remote",
	ArgsUsage: "<reference>",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "strict",
			Usage: "pull the signature without lookup the manifest",
		},
		flagOutput,
		flagUsername,
		flagPassword,
		flagPlainHTTP,
	},
	Action: runPull,
}

func runPull(ctx *cli.Context) error {
	// initialize
	if !ctx.Args().Present() {
		return errors.New("no reference specified")
	}

	reference := ctx.Args().First()
	sigRepo, err := getSignatureRepository(ctx, reference)
	if err != nil {
		return err
	}

	// core process
	if ctx.Bool("strict") {
		return pullSignatureStrict(ctx, sigRepo, reference)
	}

	manifestDesc, err := getManifestDescriptorFromReference(ctx, reference)
	if err != nil {
		return err
	}

	sigDigests, err := sigRepo.Lookup(ctx.Context, manifestDesc.Digest)
	if err != nil {
		return fmt.Errorf("lookup signature failure: %v", err)
	}

	path := ctx.String(flagOutput.Name)
	for _, sigDigest := range sigDigests {
		if path != "" {
			outputPath := filepath.Join(path, sigDigest.Encoded()+config.SignatureExtension)
			sig, err := sigRepo.Get(ctx.Context, sigDigest)
			if err != nil {
				return fmt.Errorf("get signature failure: %v: %v", sigDigest, err)
			}
			if err := osutil.WriteFile(outputPath, sig); err != nil {
				return fmt.Errorf("fail to write signature: %v: %v", sigDigest, err)
			}
		} else if err := cache.PullSignature(ctx.Context, sigRepo, manifestDesc.Digest, sigDigest); err != nil {
			return err
		}

		// write out
		fmt.Println(sigDigest)
	}

	return nil
}

func pullSignatureStrict(ctx *cli.Context, sigRepo notationregistry.SignatureRepository, reference string) error {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return err
	}
	sigDigest, err := ref.Digest()
	if err != nil {
		return fmt.Errorf("invalid signature digest: %v", err)
	}

	sig, err := sigRepo.Get(ctx.Context, sigDigest)
	if err != nil {
		return fmt.Errorf("get signature failure: %v: %v", sigDigest, err)
	}
	outputPath := ctx.String(flagOutput.Name)
	if outputPath == "" {
		outputPath = sigDigest.Encoded() + config.SignatureExtension
	}
	if err := osutil.WriteFile(outputPath, sig); err != nil {
		return fmt.Errorf("fail to write signature: %v: %v", sigDigest, err)
	}

	// write out
	fmt.Println(sigDigest)
	return nil
}

func pullSignatures(ctx *cli.Context, manifestDigest digest.Digest) error {
	reference := ctx.Args().First()
	sigRepo, err := getSignatureRepository(ctx, reference)
	if err != nil {
		return err
	}

	sigDigests, err := sigRepo.Lookup(ctx.Context, manifestDigest)
	if err != nil {
		return fmt.Errorf("lookup signature failure: %v", err)
	}
	for _, sigDigest := range sigDigests {
		if err := cache.PullSignature(ctx.Context, sigRepo, manifestDigest, sigDigest); err != nil {
			return err
		}
	}
	return nil
}
