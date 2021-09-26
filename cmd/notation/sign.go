package main

import (
	"fmt"
	"time"

	"github.com/notaryproject/notation-go-lib"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/notaryproject/notation/pkg/signature"
	"github.com/opencontainers/go-digest"
	"github.com/urfave/cli/v2"
)

var signCommand = &cli.Command{
	Name:      "sign",
	Usage:     "Signs artifacts",
	ArgsUsage: "<reference>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "key",
			Aliases: []string{"k"},
			Usage:   "signing key name",
		},
		&cli.StringFlag{
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
		localFlag,
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
		outputFlag,
		&cli.BoolFlag{
			Name:  "push",
			Usage: "push after successful signing",
			Value: true,
		},
		&cli.StringFlag{
			Name:  "push-reference",
			Usage: "different remote to store signature",
		},
		usernameFlag,
		passwordFlag,
		plainHTTPFlag,
		mediaTypeFlag,
	},
	Action: runSign,
}

func runSign(ctx *cli.Context) error {
	// initialize
	signer, err := getSigner(ctx)
	if err != nil {
		return err
	}
	_ = signer

	// core process
	desc, opts, err := prepareSigningContent(ctx)
	if err != nil {
		return err
	}
	sig, err := signer.Sign(ctx.Context, desc, opts)
	if err != nil {
		return err
	}

	// write out
	path := ctx.String(outputFlag.Name)
	if path == "" {
		path = config.SignaturePath(digest.Digest(desc.Digest), digest.FromBytes(sig))
	}
	if err := osutil.WriteFile(path, sig); err != nil {
		return err
	}

	if ctx.Bool("push") {
		ref := ctx.String("push-reference")
		if ref == "" {
			ref = ctx.Args().First()
		}
		if _, err := pushSignature(ctx, ref, sig); err != nil {
			return fmt.Errorf("fail to push signature to %q: %v: %v",
				ref,
				desc.Digest,
				err,
			)
		}
	}

	fmt.Println(desc.Digest)
	return nil
}

func prepareSigningContent(ctx *cli.Context) (notation.Descriptor, notation.SignOptions, error) {
	manifestDesc, err := getManifestDescriptorFromContext(ctx)
	if err != nil {
		return notation.Descriptor{}, notation.SignOptions{}, err
	}
	return manifestDesc, notation.SignOptions{
		Expiry: time.Now().Add(ctx.Duration("expiry")),
		Metadata: notation.Metadata{
			Identity: ctx.String("reference"),
		},
	}, nil
}

func getSigner(ctx *cli.Context) (notation.Signer, error) {
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
	certPath := ctx.String("cert-file")
	if certPath == "" {
		if name := ctx.String("cert"); name != "" {
			path, err := config.ResolveCertificatePath(name)
			if err != nil {
				return nil, err
			}
			certPath = path
		}
	}

	// construct signer
	return signature.NewSignerFromFiles(keyPath, certPath)
}
