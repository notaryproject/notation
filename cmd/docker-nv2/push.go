package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/docker/distribution/manifest/schema2"
	"github.com/notaryproject/nv2/cmd/docker-nv2/config"
	"github.com/notaryproject/nv2/cmd/docker-nv2/docker"
	"github.com/opencontainers/go-digest"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/urfave/cli/v2"
)

var pushCommand = &cli.Command{
	Name:      "push",
	Usage:     "Push an image or a repository to a registry",
	ArgsUsage: "[<reference>]",
	Action:    pushImage,
}

func pushImage(ctx *cli.Context) error {
	if err := passThroughIfNotaryDisabled(ctx); err != nil {
		return err
	}

	desc, err := pushImageAndGetOCIDescriptor(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Pushing signature")
	sigDigests, err := config.SignatureDigests(desc.Digest)
	if err != nil {
		return err
	}
	if len(sigDigests) == 0 {
		return errors.New("no signatures found")
	}

	client, err := docker.GetSignatureRepository(ctx.Context, ctx.Args().First())
	if err != nil {
		return err
	}
	pushSignature := func(sigDigest digest.Digest) error {
		sigPath := config.SignaturePath(desc.Digest, sigDigest)
		sig, err := os.ReadFile(sigPath)
		if err != nil {
			return err
		}

		sigDesc, err := client.Put(ctx.Context, sig)
		if err != nil {
			return err
		}

		artifactDesc, err := client.Link(ctx.Context, desc, sigDesc)
		if err != nil {
			return err
		}
		fmt.Println("signature manifest digest:", artifactDesc.Digest, "size:", artifactDesc.Size)
		return nil
	}
	for _, sigDigest := range sigDigests {
		if err := pushSignature(sigDigest); err != nil {
			return err
		}
	}

	return nil
}

func pushImageAndGetOCIDescriptor(ctx *cli.Context) (oci.Descriptor, error) {
	args := append([]string{"push"}, ctx.Args().Slice()...)
	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return oci.Descriptor{}, err
	}
	scanner := bufio.NewScanner(io.TeeReader(stdout, os.Stdout))
	if err := cmd.Start(); err != nil {
		return oci.Descriptor{}, err
	}
	var lastLine string
	for scanner.Scan() {
		lastLine = scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		return oci.Descriptor{}, err
	}
	if err := cmd.Wait(); err != nil {
		return oci.Descriptor{}, err
	}

	parts := strings.Split(lastLine, " ")
	if len(parts) != 5 {
		return oci.Descriptor{}, fmt.Errorf("invalid docker pull result: %s", lastLine)
	}
	digest, err := digest.Parse(parts[2])
	if err != nil {
		return oci.Descriptor{}, fmt.Errorf("invalid digest: %s", lastLine)
	}
	size, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		return oci.Descriptor{}, fmt.Errorf("invalid size: %s", lastLine)
	}

	return oci.Descriptor{
		MediaType: schema2.MediaTypeManifest,
		Digest:    digest,
		Size:      size,
	}, nil
}
