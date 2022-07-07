package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/pkg/cache"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/spf13/cobra"
)

func pushCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push [reference]",
		Short: "Push signature to remote",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPush(cmd)
		},
	}
	setFlagSignature(cmd)
	setFlagUserName(cmd)
	setFlagPassword(cmd)
	setFlagPlainHTTP(cmd)
	return cmd
}

func runPush(command *cobra.Command) error {
	// initialize
	if command.Flags().NArg() == 0 {
		return errors.New("no reference specified")
	}
	ref := command.Flags().Arg(0)
	manifestDesc, err := getManifestDescriptorFromReference(command, ref)
	if err != nil {
		return err
	}
	sigPaths, _ := command.Flags().GetStringSlice(flagSignature.Name)
	if len(sigPaths) == 0 {
		sigDigests, err := cache.SignatureDigests(manifestDesc.Digest)
		if err != nil {
			return err
		}
		for _, sigDigest := range sigDigests {
			sigPaths = append(sigPaths, config.SignaturePath(manifestDesc.Digest, sigDigest))
		}
	}

	// core process
	sigRepo, err := getSignatureRepository(command, ref)
	if err != nil {
		return err
	}
	for _, path := range sigPaths {
		sig, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		// pass in nonempty annotations if needed
		sigDesc, _, err := sigRepo.PutSignatureManifest(command.Context(), sig, manifestDesc, make(map[string]string))
		if err != nil {
			return fmt.Errorf("put signature manifest failure: %v", err)
		}

		// write out
		fmt.Println(sigDesc.Digest)
	}

	return nil
}

func pushSignature(cmd *cobra.Command, ref string, sig []byte) (notation.Descriptor, error) {
	// initialize
	sigRepo, err := getSignatureRepository(cmd, ref)
	if err != nil {
		return notation.Descriptor{}, err
	}
	manifestDesc, err := getManifestDescriptorFromReference(cmd, ref)
	if err != nil {
		return notation.Descriptor{}, err
	}

	// core process
	// pass in nonempty annotations if needed
	sigDesc, _, err := sigRepo.PutSignatureManifest(cmd.Context(), sig, manifestDesc, make(map[string]string))
	if err != nil {
		return notation.Descriptor{}, fmt.Errorf("put signature manifest failure: %v", err)
	}

	return sigDesc, nil
}
