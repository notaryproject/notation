package main

import (
	"errors"
	"fmt"
	"path/filepath"

	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/pkg/cache"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

func pullCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull [reference]",
		Short: "Pull signatures from remote",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPull(cmd)
		},
	}
	cmd.Flags().Bool("strict", false, "pull the signature without lookup the manifest")
	setFlagOutput(cmd)
	setFlagUserName(cmd)
	setFlagPassword(cmd)
	setFlagPlainHTTP(cmd)
	return cmd
}

func runPull(command *cobra.Command) error {
	// initialize
	if command.Flags().NArg() == 0 {
		return errors.New("no reference specified")
	}

	reference := command.Flags().Arg(0)
	sigRepo, err := getSignatureRepository(command, reference)
	if err != nil {
		return err
	}

	// core process
	if strict, _ := command.Flags().GetBool("strict"); strict {
		return pullSignatureStrict(command, sigRepo, reference)
	}

	manifestDesc, err := getManifestDescriptorFromReference(command, reference)
	if err != nil {
		return err
	}

	sigManifests, err := sigRepo.ListSignatureManifests(command.Context(), manifestDesc.Digest)
	if err != nil {
		return fmt.Errorf("list signature manifests failure: %v", err)
	}

	path, _ := command.Flags().GetString(flagOutput.Name)
	for _, sigManifest := range sigManifests {
		sigDigest := sigManifest.Blob.Digest
		if path != "" {
			outputPath := filepath.Join(path, sigDigest.Encoded()+config.SignatureExtension)
			sig, err := sigRepo.Get(command.Context(), sigDigest)
			if err != nil {
				return fmt.Errorf("get signature failure: %v: %v", sigDigest, err)
			}
			if err := osutil.WriteFile(outputPath, sig); err != nil {
				return fmt.Errorf("fail to write signature: %v: %v", sigDigest, err)
			}
		} else if err := cache.PullSignature(command.Context(), sigRepo, manifestDesc.Digest, sigDigest); err != nil {
			return err
		}

		// write out
		fmt.Println(sigDigest)
	}

	return nil
}

func pullSignatureStrict(command *cobra.Command, sigRepo notationregistry.SignatureRepository, reference string) error {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return err
	}
	sigDigest, err := ref.Digest()
	if err != nil {
		return fmt.Errorf("invalid signature digest: %v", err)
	}

	sig, err := sigRepo.Get(command.Context(), sigDigest)
	if err != nil {
		return fmt.Errorf("get signature failure: %v: %v", sigDigest, err)
	}
	outputPath, _ := command.Flags().GetString(flagOutput.Name)
	if outputPath == "" {
		outputPath = sigDigest.Encoded() + config.SignatureExtension
	}
	if err := osutil.WriteFile(outputPath, sig); err != nil {
		return fmt.Errorf("fail to write signature: %v: %v", sigDigest, err)
	}

	// write out
	fmt.Println(sigDigest)
	return nil
}

func pullSignatures(command *cobra.Command, manifestDigest digest.Digest) error {
	reference := command.Flags().Arg(0)
	sigRepo, err := getSignatureRepository(command, reference)
	if err != nil {
		return err
	}

	sigManifests, err := sigRepo.ListSignatureManifests(command.Context(), manifestDigest)
	if err != nil {
		return fmt.Errorf("lookup signature failure: %v", err)
	}
	for _, sigManifest := range sigManifests {
		if err := cache.PullSignature(command.Context(), sigRepo, manifestDigest, sigManifest.Blob.Digest); err != nil {
			return err
		}
	}
	return nil
}
