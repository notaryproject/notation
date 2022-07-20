package main

import (
	"errors"
	"fmt"

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

	sigManifests, err := sigRepo.ListSignatureManifests(ctx.Context, manifestDesc.Digest)
	if err != nil {
		return fmt.Errorf("lookup signature failure: %v", err)
	}

	// write out
	for _, sigManifest := range sigManifests {
		fmt.Println(sigManifest.Blob.Digest)
	}

	return nil
}
