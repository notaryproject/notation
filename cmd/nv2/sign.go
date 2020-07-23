package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/notaryproject/nv2/pkg/signature"
	"github.com/notaryproject/nv2/pkg/signature/gpg"
	"github.com/notaryproject/nv2/pkg/signature/x509"
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
			Usage:    "siging method",
			Required: true,
		},
		&cli.StringFlag{
			Name:      "key",
			Aliases:   []string{"k"},
			Usage:     "siging key file [x509]",
			TakesFile: true,
		},
		&cli.StringFlag{
			Name:      "cert",
			Aliases:   []string{"c"},
			Usage:     "siging cert [x509]",
			TakesFile: true,
		},
		&cli.StringFlag{
			Name:      "key-ring",
			Usage:     "gpg public key ring file [gpg]",
			Value:     gpg.DefaultSecretKeyRingPath(),
			TakesFile: true,
		},
		&cli.StringFlag{
			Name:      "identity",
			Aliases:   []string{"i"},
			Usage:     "signer identity [gpg]",
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
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "write signature to a specific path",
		},
		usernameFlag,
		passwordFlag,
		insecureFlag,
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
	content, err := prepareContentForSigning(ctx)
	if err != nil {
		return err
	}
	sig, err := scheme.Sign(signerID, content)
	if err != nil {
		return err
	}
	sigma, err := signature.Pack(content, sig)
	if err != nil {
		return err
	}

	// write out
	sigmaJSON, err := json.Marshal(sigma)
	if err != nil {
		return err
	}
	path := ctx.String("output")
	if path == "" {
		path = strings.Split(content.Manifests[0].Digest, ":")[1] + ".nv2"
	}
	if err := ioutil.WriteFile(path, sigmaJSON, 0666); err != nil {
		return err
	}

	fmt.Println(content.Manifests[0].Digest)
	return nil
}

func prepareContentForSigning(ctx *cli.Context) (signature.Content, error) {
	manifest, err := getManifestFromContext(ctx)
	if err != nil {
		return signature.Content{}, err
	}
	manifest.References = ctx.StringSlice("reference")
	now := time.Now()
	nowUnix := now.Unix()
	content := signature.Content{
		IssuedAt: nowUnix,
		Manifests: []signature.Manifest{
			manifest,
		},
	}
	if expiry := ctx.Duration("expiry"); expiry != 0 {
		content.NotBefore = nowUnix
		content.Expiration = now.Add(expiry).Unix()
	}

	return content, nil
}

func getSchemeForSigning(ctx *cli.Context) (*signature.Scheme, error) {
	var (
		signer signature.Signer
		err    error
	)
	switch method := ctx.String("method"); method {
	case "x509":
		signer, err = x509.NewSignerFromFiles(ctx.String("key"), ctx.String("cert"))
	case "gpg":
		signer, err = gpg.NewSigner(ctx.String("key-ring"), ctx.String("identity"))
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
