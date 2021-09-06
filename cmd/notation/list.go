package main

import (
	"errors"
	"fmt"

	"github.com/notaryproject/notation/pkg/registry"
	"github.com/urfave/cli/v2"
)

var listCommand = &cli.Command{
	Name:      "list",
	Usage:     "List signatures from remote",
	Aliases:   []string{"ls"},
	ArgsUsage: "<reference>",
	Flags: []cli.Flag{
		usernameFlag,
		passwordFlag,
		plainHTTPFlag,
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
	manifest, err := getManifestsFromReference(ctx, reference)
	if err != nil {
		return err
	}
	manifestDesc := registry.OCIDescriptorFromNotation(manifest.Descriptor)

	sigDigests, err := sigRepo.Lookup(ctx.Context, manifestDesc.Digest)
	if err != nil {
		return fmt.Errorf("lookup signature failure: %v", err)
	}

	// write out
	for _, sigDigest := range sigDigests {
		fmt.Println(sigDigest)
	}

	return nil
}
