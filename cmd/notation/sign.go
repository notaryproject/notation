package main

import (
	"fmt"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/opencontainers/go-digest"
	"github.com/urfave/cli/v2"
)

var signCommand = &cli.Command{
	Name:      "sign",
	Usage:     "Signs artifacts",
	ArgsUsage: "<reference>",
	Flags: []cli.Flag{
		cmd.FlagKey,
		cmd.FlagKeyFile,
		cmd.FlagCertFile,
		cmd.FlagTimestamp,
		cmd.FlagExpiry,
		cmd.FlagReference,
		flagLocal,
		flagOutput,
		&cli.BoolFlag{
			Name:  "push",
			Usage: "push after successful signing",
			Value: true,
		},
		&cli.StringFlag{
			Name:  "push-reference",
			Usage: "different remote to store signature",
		},
		flagUsername,
		flagPassword,
		flagPlainHTTP,
		flagMediaType,
	},
	Action: runSign,
}

func runSign(ctx *cli.Context) error {
	// initialize
	signer, err := cmd.GetSigner(ctx)
	if err != nil {
		return err
	}

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
	path := ctx.String(flagOutput.Name)
	if path == "" {
		path = config.SignaturePath(digest.Digest(desc.Digest), digest.FromBytes(sig))
	}
	if err := osutil.WriteFile(path, sig); err != nil {
		return err
	}

	if ref := ctx.String("push-reference"); ctx.Bool("push") && !(ctx.Bool(flagLocal.Name) && ref == "") {
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
	if identity := ctx.String(cmd.FlagReference.Name); identity != "" {
		manifestDesc.Annotations = map[string]string{
			"identity": identity,
		}
	}
	return manifestDesc, notation.SignOptions{
		Expiry: cmd.GetExpiry(ctx),
	}, nil
}
