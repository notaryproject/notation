package main

import (
	"fmt"

	"github.com/notaryproject/nv2/cmd/docker-nv2/config"
	"github.com/notaryproject/nv2/cmd/docker-nv2/crypto"
	"github.com/notaryproject/nv2/cmd/docker-nv2/docker"
	ios "github.com/notaryproject/nv2/internal/os"
	"github.com/opencontainers/go-digest"
	"github.com/urfave/cli/v2"
)

var notarySignCommand = &cli.Command{
	Name:      "sign",
	Usage:     "Sign a docker image",
	ArgsUsage: "[<reference>]",
	Flags: []cli.Flag{
		&cli.PathFlag{
			Name:      "key",
			Aliases:   []string{"k"},
			Usage:     "signing key file",
			TakesFile: true,
			Required:  true,
		},
		&cli.PathFlag{
			Name:      "cert",
			Aliases:   []string{"c"},
			Usage:     "signing cert",
			TakesFile: true,
		},
		&cli.StringSliceFlag{
			Name:    "reference",
			Aliases: []string{"r"},
			Usage:   "original references",
		},
		&cli.BoolFlag{
			Name:  "origin",
			Usage: "mark the current reference as a original reference",
		},
	},
	Action: notarySign,
}

func notarySign(ctx *cli.Context) error {
	if err := config.CheckNotaryEnabled(); err != nil {
		return err
	}

	service, err := crypto.GetSigningService(
		ctx.Path("key"),
		ctx.Path("cert"),
	)
	if err != nil {
		return err
	}

	reference := ctx.Args().First()
	fmt.Println("Generating Docker mainfest:", reference)
	desc, err := docker.GenerateManifestOCIDescriptor(reference)
	if err != nil {
		return err
	}

	fmt.Println("Signing", desc.Digest)
	var references []string
	if ctx.Bool("origin") {
		references = append(references, reference)
	}
	references = append(references, ctx.StringSlice("reference")...)
	sig, err := service.Sign(ctx.Context, desc, references...)
	if err != nil {
		return err
	}
	sigPath := config.SignaturePath(desc.Digest, digest.FromBytes(sig))
	if err := ios.WriteFile(sigPath, sig); err != nil {
		return err
	}
	fmt.Println("Signature saved to", sigPath)

	return nil
}
