package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/notaryproject/notation-go-lib"
	"github.com/notaryproject/notation/internal/os"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/notaryproject/notation/pkg/registry"
	"github.com/opencontainers/go-digest"
	"github.com/urfave/cli/v2"
)

var pullCommand = &cli.Command{
	Name:      "pull",
	Usage:     "Pull signatures from remote",
	ArgsUsage: "<reference>",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "strict",
			Usage: "strict pull without lookup",
		},
		outputFlag,
		usernameFlag,
		passwordFlag,
		plainHTTPFlag,
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

	manifest, err := getManifestsFromReference(ctx, reference)
	if err != nil {
		return err
	}
	manifestDesc := registry.OCIDescriptorFromNotation(manifest.Descriptor)

	sigDigests, err := sigRepo.Lookup(ctx.Context, manifestDesc.Digest)
	if err != nil {
		return fmt.Errorf("lookup signature failure: %v", err)
	}

	path := ctx.String(outputFlag.Name)
	for _, sigDigest := range sigDigests {
		sig, err := sigRepo.Get(ctx.Context, sigDigest)
		if err != nil {
			return fmt.Errorf("get signature failure: %v: %v", sigDigest, err)
		}
		var outputPath string
		if path == "" {
			outputPath = config.SignaturePath(manifestDesc.Digest, sigDigest)
		} else {
			outputPath = filepath.Join(path, sigDigest.Encoded()+config.SignatureExtension)
		}
		if err := os.WriteFile(outputPath, sig); err != nil {
			return fmt.Errorf("fail to write signature: %v: %v", sigDigest, err)
		}

		// write out
		fmt.Println(sigDigest)
	}

	return nil
}

func pullSignatureStrict(ctx *cli.Context, sigRepo notation.SignatureRepository, reference string) error {
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
	outputPath := ctx.String(outputFlag.Name)
	if outputPath == "" {
		outputPath = sigDigest.Encoded() + config.SignatureExtension
	}
	if err := os.WriteFile(outputPath, sig); err != nil {
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
		sig, err := sigRepo.Get(ctx.Context, sigDigest)
		if err != nil {
			return fmt.Errorf("get signature failure: %v: %v", sigDigest, err)
		}
		outputPath := config.SignaturePath(manifestDigest, sigDigest)
		if err := os.WriteFile(outputPath, sig); err != nil {
			return fmt.Errorf("fail to write signature: %v: %v", sigDigest, err)
		}
	}
	return nil
}
