package main

import (
	"io"
	"os"
	"os/exec"

	"github.com/notaryproject/notation/pkg/docker"
	"github.com/spf13/cobra"
)

func generateManifestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manifest [reference]",
		Short: "generates the manifest of a docker image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateManifest(cmd)
		},
	}
	cmd.Flags().StringP("output", "o", "", "write to a file instead of stdout")
	return cmd
}

func generateManifest(cmd *cobra.Command) error {
	var reader io.Reader
	if reference := cmd.Flags().Arg(0); reference != "" {
		cmd := exec.Command("docker", "save", reference)
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
	if output, _ := cmd.Flags().GetString("output"); output != "" {
		file, err := os.Create(output)
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
