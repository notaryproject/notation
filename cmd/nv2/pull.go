package main

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/notaryproject/nv2/internal/os"
	"github.com/notaryproject/nv2/pkg/registry"
	"github.com/urfave/cli/v2"
)

var pullCommand = &cli.Command{
	Name:      "pull",
	Usage:     "pull signatures from remote",
	ArgsUsage: "<scheme://reference>",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "peek",
			Usage: "view signatures without pulling",
		},
		&cli.BoolFlag{
			Name:  "strict",
			Usage: "struct pull without lookup",
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

	uri, err := url.Parse(ctx.Args().First())
	if err != nil {
		return err
	}
	sigRepo, err := getSignatureRepositoryFromURI(ctx, uri)
	if err != nil {
		return err
	}

	// core process
	if ctx.Bool("strict") {
		sigDigest, err := registry.ParseReferenceFromURL(uri).Digest()
		if err != nil {
			return fmt.Errorf("invalid signature digest: %v", err)
		}

		if !ctx.Bool("peek") {
			sig, err := sigRepo.Get(ctx.Context, sigDigest)
			if err != nil {
				return fmt.Errorf("get signature failure: %v: %v", sigDigest, err)
			}
			outputPath := ctx.String(outputFlag.Name)
			if outputPath == "" {
				outputPath = sigDigest.Encoded() + ".jwt"
			}
			if err := os.WriteFile(outputPath, sig); err != nil {
				return fmt.Errorf("fail to write signature: %v: %v", sigDigest, err)
			}
		}

		// write out
		fmt.Println(sigDigest)
		return nil
	}

	manifest, err := getManfestsFromURI(ctx, uri)
	if err != nil {
		return err
	}
	manifestDesc := registry.OCIDescriptorFromNotary(manifest.Descriptor)

	sigDigests, err := sigRepo.Lookup(ctx.Context, manifestDesc.Digest)
	if err != nil {
		return fmt.Errorf("lookup signature failure: %v", err)
	}

	path := ctx.String(outputFlag.Name)
	if path == "" {
		path = manifestDesc.Digest.Encoded()
	}
	for _, sigDigest := range sigDigests {
		if !ctx.Bool("peek") {
			sig, err := sigRepo.Get(ctx.Context, sigDigest)
			if err != nil {
				return fmt.Errorf("get signature failure: %v: %v", sigDigest, err)
			}
			outputPath := filepath.Join(path, sigDigest.Encoded()+".jwt")
			if err := os.WriteFile(outputPath, sig); err != nil {
				return fmt.Errorf("fail to write signature: %v: %v", sigDigest, err)
			}
		}

		// write out
		fmt.Println(sigDigest)
	}

	return nil
}
