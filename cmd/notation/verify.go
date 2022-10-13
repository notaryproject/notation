package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/signature"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/envelope"
	"github.com/notaryproject/notation/internal/slices"
	"github.com/notaryproject/notation/pkg/cache"
	"github.com/notaryproject/notation/pkg/configutil"
	"github.com/opencontainers/go-digest"

	"github.com/spf13/cobra"
)

type verifyOpts struct {
	RemoteFlagOpts
	signatures []string
	certs      []string
	certFiles  []string
	pull       bool
	reference  string
}

func verifyCommand(opts *verifyOpts) *cobra.Command {
	if opts == nil {
		opts = &verifyOpts{}
	}
	command := &cobra.Command{
		Use:   "verify [reference]",
		Short: "Verifies OCI Artifacts",
		Long: `
** This command is in preview and under development. **
Prerequiste: a trusted certificate needs to be generated or added using the command "notation cert". 

Example - Verify a signature using the trusted certificate:
  notation verify <registry>/<repository>:<tag>

Example - Verify a signature associated with an OCI artifact identified by the digest:
  notation verify <registry>/<repository>@<digest>

Example - Verify a signature using a trusted certificate in a specified path:
  notation verify --cert-file <cert_path> <registry>/<repository>:<tag>
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing reference")
			}
			opts.reference = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVerify(cmd, opts)
		},
	}
	setFlagSignature(command.Flags(), &opts.signatures)
	command.Flags().StringSliceVarP(&opts.certs, "cert", "c", []string{}, "certificate names for verification")
	command.Flags().StringSliceVar(&opts.certFiles, cmd.PflagCertFile.Name, []string{}, "certificate files for verification")
	command.Flags().BoolVar(&opts.pull, "pull", true, "pull remote signatures before verification")
	opts.ApplyFlags(command.Flags())
	return command
}

func runVerify(command *cobra.Command, opts *verifyOpts) error {
	// initialize
	verifier, err := getVerifier(opts)
	if err != nil {
		return err
	}
	manifestDesc, err := getManifestDescriptorFromContext(command.Context(), &opts.RemoteFlagOpts, opts.reference)
	if err != nil {
		return err
	}

	sigPaths := opts.signatures
	if len(sigPaths) == 0 {
		if !opts.Local && opts.pull {
			if err := pullSignatures(command, opts.reference, &opts.SecureFlagOpts, digest.Digest(manifestDesc.Digest)); err != nil {
				return err
			}
		}
		manifestDigest := digest.Digest(manifestDesc.Digest)
		sigDigests, err := cache.SignatureDigests(manifestDigest)
		if err != nil {
			return err
		}
		for _, sigDigest := range sigDigests {
			sigPaths = append(sigPaths, dir.Path.CachedSignature(manifestDigest, sigDigest))
		}
	}

	// core process
	if err := verifySignatures(command.Context(), verifier, manifestDesc, sigPaths); err != nil {
		return err
	}

	// write out
	fmt.Println(manifestDesc.Digest)
	return nil
}

func verifySignatures(ctx context.Context, verifier notation.Verifier, manifestDesc notation.Descriptor, sigPaths []string) error {
	if len(sigPaths) == 0 {
		return errors.New("verification failure: no signatures found")
	}

	var lastErr error
	for _, path := range sigPaths {
		sig, err := os.ReadFile(path)
		if err != nil {
			lastErr = fmt.Errorf("verification failure: %v", err)
			continue
		}
		// pass in nonempty annotations if needed
		sigMediaType, err := envelope.SpeculateSignatureEnvelopeFormat(sig)
		if err != nil {
			lastErr = fmt.Errorf("verification failure: %v", err)
			continue
		}
		opts := notation.VerifyOptions{
			SignatureMediaType: sigMediaType,
		}
		desc, err := verifier.Verify(ctx, sig, opts)
		if err != nil {
			lastErr = fmt.Errorf("verification failure: %v", err)
			continue
		}

		if !desc.Equal(manifestDesc) {
			lastErr = fmt.Errorf("verification failure: %s", manifestDesc.Digest)
			continue
		}
		return nil
	}
	return lastErr
}

func getVerifier(opts *verifyOpts) (notation.Verifier, error) {
	certPaths, err := appendCertPathFromName(opts.certFiles, opts.certs)
	if err != nil {
		return nil, err
	}
	if len(certPaths) == 0 {
		cfg, err := configutil.LoadConfigOnce()
		if err != nil {
			return nil, err
		}
		if len(cfg.VerificationCertificates.Certificates) == 0 {
			return nil, errors.New("trust certificate not specified")
		}
		for _, ref := range cfg.VerificationCertificates.Certificates {
			certPaths = append(certPaths, ref.Path)
		}
	}
	return signature.NewVerifierFromFiles(certPaths)
}

func appendCertPathFromName(paths, names []string) ([]string, error) {
	for _, name := range names {
		cfg, err := configutil.LoadConfigOnce()
		if err != nil {
			return nil, err
		}
		idx := slices.Index(cfg.VerificationCertificates.Certificates, name)
		if idx < 0 {
			return nil, errors.New("verification certificate not found: " + name)
		}
		paths = append(paths, cfg.VerificationCertificates.Certificates[idx].Path)
	}
	return paths, nil
}
