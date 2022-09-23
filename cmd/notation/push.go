package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/signature"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/pkg/cache"
	"github.com/spf13/cobra"
)

type pushOpts struct {
	SecureFlagOpts
	reference  string
	signatures []string
}

func pushCommand(opts *pushOpts) *cobra.Command {
	if opts == nil {
		opts = &pushOpts{}
	}
	cmd := &cobra.Command{
		Use:   "push [reference]",
		Short: "Push signature to remote",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no reference specified")
			}
			opts.reference = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPush(cmd, opts)
		},
	}
	setFlagSignature(cmd.Flags(), &opts.signatures)
	opts.ApplyFlags(cmd.Flags())
	return cmd
}

func runPush(command *cobra.Command, opts *pushOpts) error {
	// initialize
	ref := opts.reference
	manifestDesc, err := getManifestDescriptorFromReference(command.Context(), &opts.SecureFlagOpts, ref)
	if err != nil {
		return err
	}
	sigPaths := opts.signatures
	if len(sigPaths) == 0 {
		sigDigests, err := cache.SignatureDigests(manifestDesc.Digest)
		if err != nil {
			return err
		}
		for _, sigDigest := range sigDigests {
			sigPaths = append(sigPaths, dir.Path.CachedSignature(manifestDesc.Digest, sigDigest))
		}
	}

	// core process
	sigRepo, err := getSignatureRepository(&opts.SecureFlagOpts, ref)
	if err != nil {
		return err
	}
	for _, path := range sigPaths {
		sig, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		// pass in nonempty annotations if needed
		// TODO: understand media type in a better way
		sigMediaType, err := envelope.SpeculateSignatureEnvelopeFormat(sig)
		if err != nil {
			return err
		}
		sigDesc, _, err := sigRepo.PutSignatureManifest(command.Context(), sig, sigMediaType, manifestDesc, make(map[string]string))
		if err != nil {
			return fmt.Errorf("put signature manifest failure: %v", err)
		}

		// write out
		fmt.Println(sigDesc.Digest)
	}

	return nil
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
	// TODO: understand media type in a better way
	sigMediaType, err := signature.GuessSignatureEnvelopeFormat(sig)
	if err != nil {
		return notation.Descriptor{}, err
	}
	sigDesc, _, err := sigRepo.PutSignatureManifest(ctx, sig, sigMediaType, manifestDesc, make(map[string]string))
	if err != nil {
		return notation.Descriptor{}, fmt.Errorf("put signature manifest failure: %v", err)
	}

	return sigDesc, nil
}
