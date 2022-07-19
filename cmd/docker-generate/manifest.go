package main

import (
	"io"
	"os"
	"os/exec"

	"github.com/notaryproject/notation/pkg/docker"
	"github.com/spf13/cobra"
)

type generateManifestOpts struct {
	output    string
	reference string
}

func generateManifestCommand(opts *generateManifestOpts) *cobra.Command {
	if opts == nil {
		opts = &generateManifestOpts{}
	}
	cmd := &cobra.Command{
		Use:   "manifest [reference]",
		Short: "generates the manifest of a docker image",
		Args:  cobra.MaximumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				opts.reference = args[0]
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateManifest(cmd, opts)
		},
	}
	cmd.Flags().StringVarP(&opts.output, "output", "o", "", "write to a file instead of stdout")
	return cmd
}

func generateManifest(cmd *cobra.Command, opts *generateManifestOpts) error {
	var reader io.Reader
	if opts.reference != "" {
		cmd := exec.Command("docker", "save", opts.reference)
		cmd.Stderr = os.Stderr
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		reader = stdout
		if err := cmd.Start(); err != nil {
			return err
		}
	} else {
		reader = os.Stdin
	}

	var writer io.Writer
	if opts.output != "" {
		file, err := os.Create(opts.output)
		if err != nil {
			return err
		}
		defer file.Close()
		writer = file
	} else {
		writer = os.Stdout
	}

	manifest, err := docker.GenerateSchema2FromDockerSave(reader)
	if err != nil {
		return err
	}
	_, payload, err := manifest.Payload()
	if err != nil {
		return err
	}

	_, err = writer.Write(payload)
	return err
}
