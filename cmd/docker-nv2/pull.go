package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/docker/distribution/reference"
	"github.com/notaryproject/notary/v2"
	"github.com/notaryproject/nv2/cmd/docker-nv2/config"
	"github.com/notaryproject/nv2/cmd/docker-nv2/docker"
	ios "github.com/notaryproject/nv2/internal/os"
	"github.com/opencontainers/go-digest"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/urfave/cli/v2"
)

var pullCommand = &cli.Command{
	Name:      "pull",
	Usage:     "Pull an image or a repository from a registry",
	ArgsUsage: "[<reference>]",
	Action:    pullImage,
}

func pullImage(ctx *cli.Context) error {
	if err := passThroughIfNotaryDisabled(ctx); err != nil {
		return err
	}

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
	named, err := reference.ParseNamed(ref)
	if err != nil {
		return "", err
	}
	hostname, repository := reference.SplitHostname(named)
	manifestRef := docker.GetManifestReference(ref)

	service, err := getVerificationService()
	if err != nil {
		return "", err
	}

	manifestDesc, err := docker.GetManifestOCIDescriptor(
		ctx,
		hostname,
		repository,
		manifestRef,
	)
	if err != nil {
		return "", err
	}
	fmt.Println(manifestRef, "digest:", manifestDesc.Digest, "size:", manifestDesc.Size)

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

	sigDigest, originRefs, err := verifySignatures(ctx, service, sigDigests, manifestDesc)
	if err != nil {
		return "", fmt.Errorf("none of the signatures are valid: %v", err)
	}
	fmt.Println("Found valid signature:", sigDigest)
	if len(originRefs) != 0 {
		fmt.Println("The image is originated from:")
		for _, origin := range originRefs {
			fmt.Println("-", origin)
		}
	}

	return fmt.Sprintf("%s@%s", named.Name(), manifestDesc.Digest), nil
}

func downloadSignatures(ctx context.Context, ref string, manifestDigest digest.Digest) ([]digest.Digest, error) {
	client, err := docker.GetSignatureRepository(ctx, ref)
	if err != nil {
		return nil, err
	}
	sigDigests, err := client.Lookup(ctx, manifestDigest)
	if err != nil {
		return nil, err
	}

	for _, sigDigest := range sigDigests {
		sigPath := config.SignaturePath(manifestDigest, sigDigest)
		if _, err := os.Stat(sigPath); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return nil, err
		}

		sig, err := client.Get(ctx, sigDigest)
		if err != nil {
			return nil, err
		}
		if err := ios.WriteFile(sigPath, sig); err != nil {
			return nil, err
		}
	}

	return sigDigests, nil
}

func verifySignatures(
	ctx context.Context,
	service notary.SigningService,
	digests []digest.Digest,
	desc oci.Descriptor,
) (digest.Digest, []string, error) {
	var lastError error
	for _, digest := range digests {
		path := config.SignaturePath(desc.Digest, digest)
		sig, err := os.ReadFile(path)
		if err != nil {
			return "", nil, err
		}

		references, err := service.Verify(ctx, desc, sig)
		if err != nil {
			lastError = err
			continue
		}
		return digest, references, nil
	}
	return "", nil, lastError
}
