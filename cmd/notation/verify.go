package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/signature"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/slices"
	"github.com/notaryproject/notation/pkg/cache"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
)

func verifyCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "verify [reference]",
		Short: "Verifies OCI Artifacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVerify(cmd)
		},
	}
	setFlagSignature(command)
	command.Flags().StringSliceP("cert", "c", []string{}, "certificate names for verification")
	command.Flags().StringSlice(cmd.FlagCertFile.Name, []string{}, "certificate files for verification")
	command.Flags().Bool("pull", true, "pull remote signatures before verification")
	setFlagLocal(command)
	setFlagUserName(command)
	setFlagPassword(command)
	setFlagPlainHTTP(command)
	setFlagMediaType(command)
	return command
}

func runVerify(command *cobra.Command) error {
	// initialize
	verifier, err := getVerifier(command)
	if err != nil {
		return err
	}
	manifestDesc, err := getManifestDescriptorFromContext(command)
	if err != nil {
		return err
	}

	sigPaths, _ := command.Flags().GetStringSlice(flagSignature.Name)
	if len(sigPaths) == 0 {
		local, _ := command.Flags().GetBool(flagLocal.Name)
		if pull, _ := command.Flags().GetBool("pull"); !local && pull {
			if err := pullSignatures(command, digest.Digest(manifestDesc.Digest)); err != nil {
				return err
			}
		}
		manifestDigest := digest.Digest(manifestDesc.Digest)
		sigDigests, err := cache.SignatureDigests(manifestDigest)
		if err != nil {
			return err
		}
		for _, sigDigest := range sigDigests {
			sigPaths = append(sigPaths, config.SignaturePath(manifestDigest, sigDigest))
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

	var opts notation.VerifyOptions
	var lastErr error
	for _, path := range sigPaths {
		sig, err := os.ReadFile(path)
		if err != nil {
			return err
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

func getVerifier(command *cobra.Command) (notation.Verifier, error) {
	certPaths, _ := command.Flags().GetStringSlice(cmd.FlagCertFile.Name)
	certs, _ := command.Flags().GetStringSlice("cert")
	certPaths, err := appendCertPathFromName(certPaths, certs)
	if err != nil {
		return nil, err
	}
	if len(certPaths) == 0 {
		cfg, err := config.LoadOrDefaultOnce()
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
		cfg, err := config.LoadOrDefaultOnce()
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
