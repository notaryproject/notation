package main

import (
	"fmt"
	"time"

	"github.com/notaryproject/notation-go-lib/signature"
	"github.com/notaryproject/notation-go-lib/signature/x509"
	"github.com/notaryproject/notation/internal/os"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
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
		},
		&cli.StringSliceFlag{
			Name:    "reference",
			Aliases: []string{"r"},
			Usage:   "original references",
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
	scheme, err := getSchemeForSigning(ctx)
	if err != nil {
		return err
	}

	// core process
	claims, err := prepareClaimsForSigning(ctx)
	if err != nil {
		return err
	}
	sig, err := scheme.Sign("", claims)
	if err != nil {
		return err
	}

	// write out
	path := ctx.String(outputFlag.Name)
	if path == "" {
		path = config.SignaturePath(digest.Digest(claims.Manifest.Digest), digest.FromString(sig))
	}
	if err := os.WriteFile(path, []byte(sig)); err != nil {
		return err
	}

	if ctx.Bool("push") {
		ref := ctx.String("push-reference")
		if ref == "" {
			ref = ctx.Args().First()
		}
		if _, err := pushSignature(ctx, ref, []byte(sig)); err != nil {
			return fmt.Errorf("fail to push signature to %q: %v: %v",
				ref,
				claims.Manifest.Digest,
				err,
			)
		}
	}

	fmt.Println(claims.Manifest.Digest)
	return nil
}

func prepareClaimsForSigning(ctx *cli.Context) (signature.Claims, error) {
	manifestDesc, err := getManifestDescriptorFromContext(ctx)
	if err != nil {
		return signature.Claims{}, err
	}
	now := time.Now()
	nowUnix := now.Unix()
	claims := signature.Claims{
		Manifest: signature.Manifest{
			Descriptor: convertDescriptorToNotation(manifestDesc),
			References: ctx.StringSlice("reference"),
		},
		IssuedAt: nowUnix,
	}
	if expiry := ctx.Duration("expiry"); expiry != 0 {
		claims.NotBefore = nowUnix
		claims.Expiration = now.Add(expiry).Unix()
	}

	return claims, nil
}

func getSchemeForSigning(ctx *cli.Context) (*signature.Scheme, error) {
	keyPath := ctx.String("key-file")
	if keyPath == "" {
		path, err := config.ResolveKeyPath(ctx.String("key"))
		if err != nil {
			return nil, err
		}
		keyPath = path
	}

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

	signer, err := x509.NewSignerFromFiles(keyPath, certPath)
	scheme := signature.NewScheme()
	if err != nil {
		return nil, err
	}
	scheme.RegisterSigner("", signer)
	return scheme, nil
}

func convertDescriptorToNotation(desc ocispec.Descriptor) signature.Descriptor {
	return signature.Descriptor{
		MediaType: desc.MediaType,
		Digest:    desc.Digest.String(),
		Size:      desc.Size,
	}
}
