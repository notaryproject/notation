package main

import (
	"os"
	"os/exec"

	"github.com/notaryproject/notation-go-lib"
	"github.com/notaryproject/notation/cmd/docker-notation/crypto"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/urfave/cli/v2"
)

func passThroughIfNotationDisabled(ctx *cli.Context) error {
	err := config.CheckNotationEnabled()
	if err == nil {
		return nil
	}
	if err != config.ErrNotationDisabled {
		return err
	}

	args := append([]string{ctx.Command.Name}, ctx.Args().Slice()...)
	if err := runCommand("docker", args...); err != nil {
		return err
	}
	os.Exit(0)
	panic("process should be terminated")
}

func runCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
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

func getVerificationService() (notation.SigningService, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	var certPaths []string
	for _, cert := range cfg.VerificationCertificates.Certificates {
		certPaths = append(certPaths, cert.Path)
	}
	return crypto.GetSigningService("", certPaths...)
}
