package main

import (
	"fmt"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/cmd/docker-notation/docker"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
)

func signCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "sign [reference]",
		Short: "Sign a image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return signImage(cmd)
		},
	}
	cmd.SetFlagKey(command)
	cmd.SetFlagKeyFile(command)
	cmd.SetFlagCertFile(command)
	cmd.SetFlagTimestamp(command)
	cmd.SetFlagExpiry(command)
	cmd.SetFlagReference(command)

	command.Flags().Bool("origin", false, "mark the current reference as a original reference")
	return command
}

func signImage(command *cobra.Command) error {
	signer, err := cmd.GetSigner(command)
	if err != nil {
		return err
	}

	reference := command.Flags().Arg(0)
	fmt.Println("Generating Docker mainfest:", reference)
	desc, err := docker.GenerateManifestDescriptor(reference)
	if err != nil {
		return err
	}

	fmt.Println("Signing", desc.Digest)

	identity, _ := command.Flags().GetString(cmd.FlagReference.Name)
	if origin, _ := command.Flags().GetBool("origin"); origin {
		identity = reference
	}
	if identity != "" {
		desc.Annotations = map[string]string{
			"identity": identity,
		}
	}
	sig, err := signer.Sign(command.Context(), desc, notation.SignOptions{
		Expiry: cmd.GetExpiry(command),
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
