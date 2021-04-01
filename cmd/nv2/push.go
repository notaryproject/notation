package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/notaryproject/nv2/pkg/registry"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
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
	uri := ctx.Args().First()
	sig, err := os.ReadFile(ctx.String("signature"))
	if err != nil {
		return err
	}

	// core process
	desc, err := pushSignature(ctx, uri, sig)
	if err != nil {
		return err
	}

	// write out
	fmt.Println(desc.Digest)
	return nil
}

func pushSignature(ctx *cli.Context, uri string, sig []byte) (oci.Descriptor, error) {
	// initialize
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return oci.Descriptor{}, err
	}
	sigRepo, err := getSignatureRepositoryFromURI(ctx, parsedURI)
	if err != nil {
		return oci.Descriptor{}, err
	}
	manifest, err := getManfestsFromURI(ctx, parsedURI)
	if err != nil {
		return oci.Descriptor{}, err
	}
	manifestDesc := registry.OCIDescriptorFromNotary(manifest.Descriptor)

	// core process
	sigDesc, err := sigRepo.Put(ctx.Context, sig)
	if err != nil {
		return oci.Descriptor{}, fmt.Errorf("push signature failure: %v", err)
	}

	desc, err := sigRepo.Link(ctx.Context, manifestDesc, sigDesc)
	if err != nil {
		return oci.Descriptor{}, fmt.Errorf("link signature failure: %v", err)
	}

	return desc, nil
}
