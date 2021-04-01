package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/notaryproject/nv2/pkg/registry"
	"github.com/urfave/cli/v2"
)

var pushCommand = &cli.Command{
	Name:      "push",
	Usage:     "push signature to remote",
	ArgsUsage: "<scheme://reference>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:      "signature",
			Aliases:   []string{"s", "f"},
			Usage:     "signature file",
			Required:  true,
			TakesFile: true,
		},
		usernameFlag,
		passwordFlag,
		plainHTTPFlag,
	},
	Action: runPush,
}

func runPush(ctx *cli.Context) error {
	// initialize
	if !ctx.Args().Present() {
		return errors.New("no reference specified")
	}
	uri, err := url.Parse(ctx.Args().First())
	if err != nil {
		return err
	}
	sig, err := os.ReadFile(ctx.String("signature"))
	if err != nil {
		return err
	}
	sigRepo, err := getSignatureRepositoryFromURI(ctx, uri)
	if err != nil {
		return err
	}
	manifest, err := getManfestsFromURI(ctx, uri)
	if err != nil {
		return err
	}
	manifestDesc := registry.OCIDescriptorFromNotary(manifest.Descriptor)

	// core process
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
	return nil
}
