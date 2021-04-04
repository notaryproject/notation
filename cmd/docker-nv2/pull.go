package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/distribution/reference"
	"github.com/notaryproject/notary/v2"
	"github.com/notaryproject/nv2/cmd/docker-nv2/docker"
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
	client, err := docker.GetSignatureRepository(ctx, ref)
	if err != nil {
		return "", err
	}
	sigDigests, err := client.Lookup(ctx, manifestDesc.Digest)
	if err != nil {
		return "", err
	}
	switch n := len(sigDigests); n {
	case 0:
		return "", errors.New("no signature found")
	default:
		fmt.Println("Found", n, "signatures")
	}

	sigDigest, originRefs, err := verifySignatures(
		ctx,
		service,
		client,
		sigDigests,
		manifestDesc,
	)
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

func verifySignatures(
	ctx context.Context,
	service notary.SigningService,
	client notary.SignatureRepository,
	digests []digest.Digest,
	desc oci.Descriptor,
) (digest.Digest, []string, error) {
	var lastError error
	for _, digest := range digests {
		sig, err := client.Get(ctx, digest)
		if err != nil {
			lastError = err
			continue
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
