package main

import (
	"fmt"

	"github.com/notaryproject/notation-go-lib"
	"github.com/notaryproject/notation/cmd/docker-notation/crypto"
	"github.com/notaryproject/notation/cmd/docker-notation/docker"
	ios "github.com/notaryproject/notation/internal/os"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/opencontainers/go-digest"
	"github.com/urfave/cli/v2"
)

var signCommand = &cli.Command{
	Name:      "sign",
	Usage:     "Sign a image",
	ArgsUsage: "<reference>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "key",
			Aliases: []string{"k"},
			Usage:   "signing key name",
		},
		&cli.PathFlag{
			Name:      "key-file",
			Usage:     "signing key file",
			TakesFile: true,
		},
		&cli.StringFlag{
			Name:    "cert",
			Aliases: []string{"c"},
			Usage:   "signing certificate name",
		},
		&cli.StringFlag{
			Name:      "cert-file",
			Usage:     "signing certificate file",
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
	Action: signImage,
}

func signImage(ctx *cli.Context) error {
	service, err := getSigningService(ctx)
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

func getSigningService(ctx *cli.Context) (notation.SigningService, error) {
	keyPath := ctx.String("key-file")
	if keyPath == "" {
		path, err := config.ResolveKeyPath(ctx.String("key"))
		if err != nil {
			return nil, err
		}
		keyPath = path
	}

	var certPaths []string
	if path := ctx.String("cert-file"); path != "" {
		certPaths = []string{path}
	} else if name := ctx.String("cert"); name != "" {
		path, err := config.ResolveCertificatePath(name)
		if err != nil {
			return nil, err
		}
		certPaths = []string{path}
	}

	return crypto.GetSigningService(keyPath, certPaths...)
}
