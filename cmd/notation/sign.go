package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/crypto/timestamp"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
)

type signOpts struct {
	cmd.SignerFlagOpts
	RemoteFlagOpts
	timestamp       string
	expiry          time.Duration
	originReference string
	output          string
	push            bool
	pushReference   string
	pluginConfig    string
	reference       string
}

func signCommand(opts *signOpts) *cobra.Command {
	if opts == nil {
		opts = &signOpts{}
	}
	command := &cobra.Command{
		Use:   "sign [reference]",
		Short: "Signs artifacts",
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
	opts.RemoteFlagOpts.ApplyFlags(command.Flags())

	cmd.SetPflagTimestamp(command.Flags(), &opts.timestamp)
	cmd.SetPflagExpiry(command.Flags(), &opts.expiry)
	cmd.SetPflagReference(command.Flags(), &opts.originReference)
	setFlagOutput(command.Flags(), &opts.output)

	command.Flags().BoolVar(&opts.push, "push", true, "push after successful signing")
	command.Flags().StringVar(&opts.pushReference, "push-reference", "", "different remote to store signature")

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
	path := cmdOpts.output
	if path == "" {
		path = config.SignaturePath(digest.Digest(desc.Digest), digest.FromBytes(sig))
	}
	if err := osutil.WriteFile(path, sig); err != nil {
		return err
	}

	if ref := cmdOpts.pushReference; cmdOpts.push && !(cmdOpts.Local && ref == "") {
		if ref == "" {
			ref = cmdOpts.reference
		}
		if _, err := pushSignature(command.Context(), &cmdOpts.SecureFlagOpts, ref, sig); err != nil {
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

func prepareSigningContent(ctx context.Context, opts *signOpts) (notation.Descriptor, notation.SignOptions, error) {
	manifestDesc, err := getManifestDescriptorFromContext(ctx, &opts.RemoteFlagOpts, opts.reference)
	if err != nil {
		return notation.Descriptor{}, notation.SignOptions{}, err
	}
	if identity := opts.originReference; identity != "" {
		manifestDesc.Annotations = map[string]string{
			"identity": identity,
		}
	}
	var tsa timestamp.Timestamper
	if endpoint := opts.timestamp; endpoint != "" {
		if tsa, err = timestamp.NewHTTPTimestamper(nil, endpoint); err != nil {
			return notation.Descriptor{}, notation.SignOptions{}, err
		}
	}
	pluginConfig, err := cmd.ParseFlagPluginConfig(opts.pluginConfig)
	if err != nil {
		return notation.Descriptor{}, notation.SignOptions{}, err
	}
	return manifestDesc, notation.SignOptions{
		Expiry:       cmd.GetExpiry(opts.expiry),
		TSA:          tsa,
		PluginConfig: pluginConfig,
	}, nil
}
