package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/pkg/cache"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/urfave/cli/v2"
)

var pushCommand = &cli.Command{
	Name:      "push",
	Usage:     "Push signature to remote",
	ArgsUsage: "<reference>",
	Flags: []cli.Flag{
		flagSignature,
		flagUsername,
		flagPassword,
		flagPlainHTTP,
	},
	Action: runPush,
}

func runPush(ctx *cli.Context) error {
	// initialize
	if !ctx.Args().Present() {
		return errors.New("no reference specified")
	}
	ref := ctx.Args().First()
	manifestDesc, err := getManifestDescriptorFromReference(ctx, ref)
	if err != nil {
		return err
	}
	sigPaths := ctx.StringSlice(flagSignature.Name)
	if len(sigPaths) == 0 {
		sigDigests, err := cache.SignatureDigests(manifestDesc.Digest)
		if err != nil {
			return err
		}
		for _, sigDigest := range sigDigests {
			sigPaths = append(sigPaths, config.SignaturePath(manifestDesc.Digest, sigDigest))
		}
	}

	// core process
	sigRepo, err := getSignatureRepository(ctx, ref)
	if err != nil {
		return err
	}
	for _, path := range sigPaths {
		sig, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sigDesc, err := sigRepo.Put(ctx.Context, sig)
		if err != nil {
			return fmt.Errorf("push signature failure: %v", err)
		}
		desc, err := sigRepo.Link(ctx.Context, manifestDesc, sigDesc)
		if err != nil {
			return fmt.Errorf("link signature failure: %v", err)
		}

		// write out
		fmt.Println(desc.Digest)
	}

	return nil
}

func pushSignature(ctx *cli.Context, ref string, sig []byte) (notation.Descriptor, error) {
	// initialize
	sigRepo, err := getSignatureRepository(ctx, ref)
	if err != nil {
		return notation.Descriptor{}, err
	}
	manifestDesc, err := getManifestDescriptorFromReference(ctx, ref)
	if err != nil {
		return notation.Descriptor{}, err
	}

	// core process
	sigDesc, err := sigRepo.Put(ctx.Context, sig)
	if err != nil {
		return notation.Descriptor{}, fmt.Errorf("push signature failure: %v", err)
	}
	desc, err := sigRepo.Link(ctx.Context, manifestDesc, sigDesc)
	if err != nil {
		return notation.Descriptor{}, fmt.Errorf("link signature failure: %v", err)
	}

	return desc, nil
}
