package main

import (
	"fmt"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/crypto/timestamp"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
)

func signCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "sign [reference]",
		Short: "Signs artifacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSign(cmd)
		},
	}
	cmd.SetFlagKey(command)
	cmd.SetFlagKeyFile(command)
	cmd.SetFlagCertFile(command)
	cmd.SetFlagTimestamp(command)
	cmd.SetFlagExpiry(command)
	cmd.SetFlagReference(command)
	setFlagLocal(command)
	setFlagOutput(command)

	command.Flags().Bool("push", true, "push after successful signing")
	command.Flags().String("push-reference", "", "different remote to store signature")

	setFlagUserName(command)
	setFlagPassword(command)
	setFlagPlainHTTP(command)
	setFlagMediaType(command)

	cmd.SetFlagPluginConfig(command)
	return command
}

func runSign(command *cobra.Command) error {
	// initialize
	signer, err := cmd.GetSigner(command)
	if err != nil {
		return err
	}

	// core process
	desc, opts, err := prepareSigningContent(command)
	if err != nil {
		return err
	}
	sig, err := signer.Sign(command.Context(), desc, opts)
	if err != nil {
		return err
	}

	// write out
	path, _ := command.Flags().GetString(flagOutput.Name)
	if path == "" {
		path = config.SignaturePath(digest.Digest(desc.Digest), digest.FromBytes(sig))
	}
	if err := osutil.WriteFile(path, sig); err != nil {
		return err
	}

	ref, _ := command.Flags().GetString("push-reference")
	push, _ := command.Flags().GetBool("push")
	if local, _ := command.Flags().GetBool(flagLocal.Name); push && !(local && ref == "") {
		if ref == "" {
			ref = command.Flags().Arg(0)
		}
		if _, err := pushSignature(command, ref, sig); err != nil {
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

func prepareSigningContent(command *cobra.Command) (notation.Descriptor, notation.SignOptions, error) {
	manifestDesc, err := getManifestDescriptorFromContext(command)
	if err != nil {
		return notation.Descriptor{}, notation.SignOptions{}, err
	}
	if identity, _ := command.Flags().GetString(cmd.FlagReference.Name); identity != "" {
		manifestDesc.Annotations = map[string]string{
			"identity": identity,
		}
	}
	var tsa timestamp.Timestamper
	if endpoint, _ := command.Flags().GetString(cmd.FlagTimestamp.Name); endpoint != "" {
		if tsa, err = timestamp.NewHTTPTimestamper(nil, endpoint); err != nil {
			return notation.Descriptor{}, notation.SignOptions{}, err
		}
	}
	pluginConfig, err := cmd.ParseFlagPluginConfig(command)
	if err != nil {
		return notation.Descriptor{}, notation.SignOptions{}, err
	}
	return manifestDesc, notation.SignOptions{
		Expiry:       cmd.GetExpiry(command),
		TSA:          tsa,
		PluginConfig: pluginConfig,
	}, nil
}
