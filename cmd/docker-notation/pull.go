package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/cmd/docker-notation/docker"
	"github.com/notaryproject/notation/pkg/cache"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/opencontainers/go-digest"
	"github.com/urfave/cli/v2"
	"oras.land/oras-go/v2/registry"
)

var pullCommand = &cli.Command{
	Name:      "pull",
	Usage:     "Verify and pull an image from a registry",
	ArgsUsage: "<reference>",
	Action:    pullImage,
}

func pullImage(ctx *cli.Context) error {
	originalRef := ctx.Args().First()
	ref, err := verifyRemoteImage(ctx.Context, originalRef)
	if err != nil {
		return err
	}

	if err := runCommand("docker", "pull", ref); err != nil {
		return err
	}
	return runCommand("docker", "tag", ref, originalRef)
}

func verifyRemoteImage(ctx context.Context, ref string) (string, error) {
	manifestRef, err := registry.ParseReference(ref)
	if err != nil {
		return "", err
	}

	verifier, err := getVerifier()
	if err != nil {
		return "", err
	}

	manifestDesc, err := docker.GetManifestDescriptor(ctx, manifestRef)
	if err != nil {
		return "", err
	}
	fmt.Printf("%s: digest: %v size: %v\n", manifestRef.ReferenceOrDefault(), manifestDesc.Digest, manifestDesc.Size)

	fmt.Println("Looking up for signatures")
	sigDigests, err := downloadSignatures(ctx, ref, manifestDesc.Digest)
	if err != nil {
		return "", err
	}
	switch n := len(sigDigests); n {
	case 0:
		return "", errors.New("no signature found")
	default:
		fmt.Println("Found", n, "signatures")
	}

	sigDigest, originRef, err := verifySignatures(ctx, verifier, sigDigests, manifestDesc)
	if err != nil {
		return "", fmt.Errorf("none of the signatures are valid: %v", err)
	}
	fmt.Println("Found valid signature:", sigDigest)
	if originRef != "" {
		fmt.Println("The image is originated from:", originRef)
	}

	manifestRef.Reference = manifestDesc.Digest.String()
	return manifestRef.String(), nil
}

func downloadSignatures(ctx context.Context, ref string, manifestDigest digest.Digest) ([]digest.Digest, error) {
	client, err := docker.GetSignatureRepository(ref)
	if err != nil {
		return nil, err
	}
	sigDigests, err := client.Lookup(ctx, manifestDigest)
	if err != nil {
		return nil, err
	}

	for _, sigDigest := range sigDigests {
		if err := cache.PullSignature(ctx, client, manifestDigest, sigDigest); err != nil {
			return nil, err
		}
	}

	return sigDigests, nil
}

func verifySignatures(
	ctx context.Context,
	verifier notation.Verifier,
	sigDigests []digest.Digest,
	desc notation.Descriptor,
) (digest.Digest, string, error) {
	var opts notation.VerifyOptions
	var lastErr error
	for _, sigDigest := range sigDigests {
		path := config.SignaturePath(desc.Digest, sigDigest)
		sig, err := os.ReadFile(path)
		if err != nil {
			return "", "", err
		}

		actualDesc, err := verifier.Verify(ctx, sig, opts)
		if err != nil {
			lastErr = err
			continue
		}
		if !actualDesc.Equal(desc) {
			lastErr = fmt.Errorf("verification failure: digest mismatch: %v: %v", desc.Digest, actualDesc.Digest)
			continue
		}
		return sigDigest, actualDesc.Annotations["identity"], nil
	}
	return "", "", lastErr
}
