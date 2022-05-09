package main

import (
	"errors"
	"fmt"

	"github.com/opencontainers/go-digest"
	"github.com/urfave/cli/v2"
)

var listCommand = &cli.Command{
	Name:      "list",
	Usage:     "List signatures from remote",
	Aliases:   []string{"ls"},
	ArgsUsage: "<reference>",
	Flags: []cli.Flag{
		flagUsername,
		flagPassword,
		flagPlainHTTP,
	},
	Action: runList,
}

func runList(ctx *cli.Context) error {
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
	manifestDesc, err := getManifestDescriptorFromReference(ctx, reference)
	if err != nil {
		return err
	}

	sigDigests, err := sigRepo.Lookup(ctx.Context, digest.Digest(manifestDesc.Digest))
	if err != nil {
		return fmt.Errorf("lookup signature failure: %v", err)
	}

	// write out
	for _, sigDigest := range sigDigests {
		fmt.Println(sigDigest)
	}

	return nil
}
