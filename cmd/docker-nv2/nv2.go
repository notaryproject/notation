package main

import (
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

var nv2Command = &cli.Command{
	Name: "nv2",
	Subcommands: []*cli.Command{
		notaryCommand,
		pullCommand,
		pushCommand,
		certsCommand,
	},
	Action: delegateToDocker,
}

func delegateToDocker(ctx *cli.Context) error {
	cmd := exec.Command("docker", ctx.Args().Slice()...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			os.Exit(err.ExitCode())
		}
		return err
	}
	return nil
}
