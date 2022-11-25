package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/spf13/cobra"
)

type signOpts struct {
	cmd.SignerFlagOpts
	SecureFlagOpts
	expiry       time.Duration
	pluginConfig []string
	reference    string
}

func signCommand(opts *signOpts) *cobra.Command {
	if opts == nil {
		opts = &signOpts{}
	}
	command := &cobra.Command{
		Use:   "sign [flags] <reference>",
		Short: "Sign artifacts",
		Long: `Sign artifacts

Prerequisite: a signing key needs to be configured using the command "notation key".

Example - Sign an OCI artifact using the default signing key, with the default JWS envelope:
  notation sign <registry>/<repository>@<digest>

Example - Sign an OCI artifact using the default signing key, with the COSE envelope:
  notation sign --signature-format cose <registry>/<repository>@<digest> 

Example - Sign an OCI artifact using a specified key
  notation sign --key <key_name> <registry>/<repository>@<digest>

Example - Sign an OCI artifact using a local key and certificate file directly
  notation sign --key-file <key_path> --cert-file <cert_path> <registry>/<repository>@<digest>

Example - Sign an OCI artifact identified by a tag (Notation will resolve tag to digest)
  notation sign <registry>/<repository>:<tag>

Example - Sign an OCI artifact stored in a registry and specify the signature expiry duration, for example 24 hours
  notation sign --expiry 24h <registry>/<repository>@<digest>
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing reference")
			}
			opts.reference = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSign(cmd, opts)
		},
	}
	opts.SignerFlagOpts.ApplyFlags(command.Flags())
	opts.SecureFlagOpts.ApplyFlags(command.Flags())
	cmd.SetPflagExpiry(command.Flags(), &opts.expiry)
	cmd.SetPflagPluginConfig(command.Flags(), &opts.pluginConfig)

	return command
}

func runSign(command *cobra.Command, cmdOpts *signOpts) error {
	// initialize
	signer, err := cmd.GetSigner(&cmdOpts.SignerFlagOpts)
	if err != nil {
		return err
	}

	// core process
	desc, opts, err := prepareSigningContent(command.Context(), cmdOpts)
	if err != nil {
		return err
	}
	sig, err := signer.Sign(command.Context(), desc, opts)
	if err != nil {
		return err
	}

	// write out
	ref := cmdOpts.reference
	if _, err := pushSignature(command.Context(), &cmdOpts.SecureFlagOpts, ref, sig); err != nil {
		return fmt.Errorf("fail to push signature to %q: %v: %v",
			ref,
			desc.Digest,
			err,
		)
	}

	fmt.Println(desc.Digest)
	return nil
}

func prepareSigningContent(ctx context.Context, opts *signOpts) (notation.Descriptor, notation.SignOptions, error) {
	manifestDesc, err := getManifestDescriptorFromContext(ctx, &opts.SecureFlagOpts, opts.reference)
	if err != nil {
		return notation.Descriptor{}, notation.SignOptions{}, err
	}
	pluginConfig, err := cmd.ParseFlagPluginConfig(opts.pluginConfig)
	if err != nil {
		return notation.Descriptor{}, notation.SignOptions{}, err
	}
	return manifestDesc, notation.SignOptions{
		Expiry:       cmd.GetExpiry(opts.expiry),
		PluginConfig: pluginConfig,
	}, nil
}

func pushSignature(ctx context.Context, opts *SecureFlagOpts, ref string, sig []byte) (notation.Descriptor, error) {
	// initialize
	sigRepo, err := getSignatureRepository(opts, ref)
	if err != nil {
		return notation.Descriptor{}, err
	}
	manifestDesc, err := getManifestDescriptorFromReference(ctx, opts, ref)
	if err != nil {
		return notation.Descriptor{}, err
	}

	// core process
	// pass in nonempty annotations if needed
	sigMediaType, err := envelope.SpeculateSignatureEnvelopeFormat(sig)
	if err != nil {
		return notation.Descriptor{}, err
	}
	sigDesc, _, err := sigRepo.PutSignatureManifest(ctx, sig, sigMediaType, manifestDesc, make(map[string]string))
	if err != nil {
		return notation.Descriptor{}, fmt.Errorf("put signature manifest failure: %v", err)
	}

	return sigDesc, nil
}
