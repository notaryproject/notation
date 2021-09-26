package main

import (
	"fmt"
	"time"

	"github.com/notaryproject/notation-go-lib"
	"github.com/notaryproject/notation/cmd/docker-notation/docker"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/notaryproject/notation/pkg/signature"
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
		&cli.DurationFlag{
			Name:    "expiry",
			Aliases: []string{"e"},
			Usage:   "expire duration",
			Value:   7 * 24 * time.Hour, // default to a week
		},
		&cli.StringFlag{
			Name:    "reference",
			Aliases: []string{"r"},
			Usage:   "original reference",
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
	desc, err := docker.GenerateManifestDescriptor(reference)
	if err != nil {
		return err
	}

	fmt.Println("Signing", desc.Digest)
	identity := ctx.String("reference")
	if ctx.Bool("origin") {
		identity = reference
	}
	sig, err := service.Sign(ctx.Context, desc, notation.SignOptions{
		Expiry: time.Now().Add(ctx.Duration("expiry")),
		Metadata: notation.Metadata{
			Identity: identity,
		},
	})
	if err != nil {
		return err
	}
	sigPath := config.SignaturePath(desc.Digest, digest.FromBytes(sig))
	if err := osutil.WriteFile(sigPath, sig); err != nil {
		return err
	}
	fmt.Println("Signature saved to", sigPath)

	return nil
}

func getSigningService(ctx *cli.Context) (notation.Signer, error) {
	// read signing key
	keyPath := ctx.String("key-file")
	if keyPath == "" {
		path, err := config.ResolveKeyPath(ctx.String("key"))
		if err != nil {
			return nil, err
		}
		keyPath = path
	}

	// read certs associated with the signing
	var certPath string
	if path := ctx.String("cert-file"); path != "" {
		certPath = path
	} else if name := ctx.String("cert"); name != "" {
		path, err := config.ResolveCertificatePath(name)
		if err != nil {
			return nil, err
		}
		certPath = path
	}

	return signature.NewSignerFromFiles(keyPath, certPath)
}
