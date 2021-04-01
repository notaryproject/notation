package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/notaryproject/notary/v2/signature"
	"github.com/notaryproject/notary/v2/signature/x509"
	"github.com/notaryproject/nv2/internal/os"
	"github.com/urfave/cli/v2"
)

const signerID = "nv2"

var signCommand = &cli.Command{
	Name:      "sign",
	Usage:     "signs OCI Artifacts",
	ArgsUsage: "[<scheme://reference>]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "method",
			Aliases:  []string{"m"},
			Usage:    "signing method",
			Required: true,
		},
		&cli.StringFlag{
			Name:      "key",
			Aliases:   []string{"k"},
			Usage:     "signing key file [x509]",
			TakesFile: true,
		},
		&cli.StringFlag{
			Name:      "cert",
			Aliases:   []string{"c"},
			Usage:     "signing cert [x509]",
			TakesFile: true,
		},
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
	sig, err := scheme.Sign(signerID, claims)
	if err != nil {
		return err
	}

	// write out
	path := ctx.String(outputFlag.Name)
	if path == "" {
		path = strings.Split(claims.Manifest.Digest, ":")[1] + ".jwt"
	}
	if err := os.WriteFile(path, []byte(sig)); err != nil {
		return err
	}

	if ctx.Bool("push") {
		uri := ctx.String("push-reference")
		if uri == "" {
			uri = ctx.Args().First()
		}
		if _, err := pushSignature(ctx, uri, []byte(sig)); err != nil {
			return fmt.Errorf("fail to push signature to %q: %v: %v",
				uri,
				claims.Manifest.Digest,
				err,
			)
		}
	}

	fmt.Println(claims.Manifest.Digest)
	return nil
}

func prepareClaimsForSigning(ctx *cli.Context) (signature.Claims, error) {
	manifest, err := getManifestFromContext(ctx)
	if err != nil {
		return signature.Claims{}, err
	}
	manifest.References = ctx.StringSlice("reference")
	now := time.Now()
	nowUnix := now.Unix()
	claims := signature.Claims{
		Manifest: manifest,
		IssuedAt: nowUnix,
	}
	if expiry := ctx.Duration("expiry"); expiry != 0 {
		claims.NotBefore = nowUnix
		claims.Expiration = now.Add(expiry).Unix()
	}

	return claims, nil
}

func getSchemeForSigning(ctx *cli.Context) (*signature.Scheme, error) {
	var (
		signer signature.Signer
		err    error
	)
	switch method := ctx.String("method"); method {
	case "x509":
		signer, err = x509.NewSignerFromFiles(ctx.String("key"), ctx.String("cert"))
	default:
		return nil, fmt.Errorf("unsupported signing method: %s", method)
	}
	scheme := signature.NewScheme()
	if err != nil {
		return nil, err
	}
	scheme.RegisterSigner(signerID, signer)
	return scheme, nil
}
