package main

import (
	"io"
	"os"
	"os/exec"

	"github.com/notaryproject/nv2/pkg/docker"
	"github.com/urfave/cli/v2"
)

var manifestCommand = &cli.Command{
	Name:      "manifest",
	Usage:     "generates the manifest of a docker image",
	ArgsUsage: "[<reference>]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "write to a file instead of stdout",
		},
	},
	Action: generateManifest,
}

func generateManifest(ctx *cli.Context) error {
	var reader io.Reader
	if reference := ctx.Args().First(); reference != "" {
		cmd := exec.Command("docker", "save", reference)
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
	if output := ctx.String("output"); output != "" {
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
