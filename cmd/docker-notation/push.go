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

	"github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/cmd/docker-notation/docker"
	"github.com/notaryproject/notation/pkg/cache"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
)

func pushCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push [reference]",
		Short: "Push an image to a registry with its signatures",
		RunE: func(cmd *cobra.Command, args []string) error {
			return pushImage(cmd)
		},
	}
	return cmd
}

func pushImage(cmd *cobra.Command) error {
	desc, err := pushImageAndGetDescriptor(cmd)
	if err != nil {
		return err
	}

	fmt.Println("Pushing signature")
	sigDigests, err := cache.SignatureDigests(desc.Digest)
	if err != nil {
		return err
	}
	if len(sigDigests) == 0 {
		return errors.New("no signatures found")
	}

	client, err := docker.GetSignatureRepository(cmd.Flags().Arg(0))
	if err != nil {
		return err
	}
	pushSignature := func(sigDigest digest.Digest) error {
		sigPath := config.SignaturePath(desc.Digest, sigDigest)
		sig, err := os.ReadFile(sigPath)
		if err != nil {
			return err
		}

		// pass in nonempty annotations if needed
		sigDesc, _, err := client.PutSignatureManifest(cmd.Context(), sig, desc, make(map[string]string))
		if err != nil {
			return err
		}
		fmt.Println("signature manifest digest:", sigDesc.Digest, "size:", sigDesc.Size)
		return nil
	}
	for _, sigDigest := range sigDigests {
		if err := pushSignature(sigDigest); err != nil {
			return err
		}
	}

	return nil
}

func pushImageAndGetDescriptor(pushCmd *cobra.Command) (notation.Descriptor, error) {
	args := append([]string{"push"}, pushCmd.Flags().Args()...)
	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return notation.Descriptor{}, err
	}
	scanner := bufio.NewScanner(io.TeeReader(stdout, os.Stdout))
	if err := cmd.Start(); err != nil {
		return notation.Descriptor{}, err
	}
	var lastLine string
	for scanner.Scan() {
		lastLine = scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		return notation.Descriptor{}, err
	}
	if err := cmd.Wait(); err != nil {
		return notation.Descriptor{}, err
	}

	parts := strings.Split(lastLine, " ")
	if len(parts) != 5 {
		return notation.Descriptor{}, fmt.Errorf("invalid docker pull result: %s", lastLine)
	}
	digest, err := digest.Parse(parts[2])
	if err != nil {
		return notation.Descriptor{}, fmt.Errorf("invalid digest: %s", lastLine)
	}
	size, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		return notation.Descriptor{}, fmt.Errorf("invalid size: %s", lastLine)
	}

	return notation.Descriptor{
		MediaType: schema2.MediaTypeManifest,
		Digest:    digest,
		Size:      size,
	}, nil
}
