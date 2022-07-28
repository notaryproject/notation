package main

import (
	"fmt"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/cmd/docker-notation/docker"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/opencontainers/go-digest"
	"github.com/urfave/cli/v2"
)

var signCommand = &cli.Command{
	Name:      "sign",
	Usage:     "Sign a image",
	ArgsUsage: "<reference>",
	Flags: []cli.Flag{
		cmd.FlagKey,
		cmd.FlagKeyFile,
		cmd.FlagCertFile,
		cmd.FlagTimestamp,
		cmd.FlagExpiry,
		cmd.FlagReference,
		&cli.BoolFlag{
			Name:  "origin",
			Usage: "mark the current reference as a original reference",
		},
	},
	Action: signImage,
}

func signImage(ctx *cli.Context) error {
	// TODO: make this change only to make sure the code can be compiled
	// According to the https://github.com/notaryproject/notation/discussions/251,
	// we can update/deprecate it later
	signerOpts := &cmd.SignerFlagOpts{
		Key:      ctx.String(cmd.FlagKey.Name),
		KeyFile:  ctx.String(cmd.FlagKeyFile.Name),
		CertFile: ctx.String(cmd.FlagCertFile.Name),
	}
	signer, err := cmd.GetSigner(signerOpts)
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
	identity := ctx.String(cmd.FlagReference.Name)
	if ctx.Bool("origin") {
		identity = reference
	}
	if identity != "" {
		desc.Annotations = map[string]string{
			"identity": identity,
		}
	}
	sig, err := signer.Sign(ctx.Context, desc, notation.SignOptions{
		Expiry: cmd.GetExpiry(ctx.Duration(cmd.FlagExpiry.Name)),
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
