package main

import (
	"os"
	"os/exec"

	"github.com/notaryproject/notation-go-lib"
	"github.com/notaryproject/notation/cmd/docker-notation/crypto"
	"github.com/notaryproject/notation/pkg/config"
)

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
	cfg, err := config.LoadOrDefaultOnce()
	if err != nil {
		return nil, err
	}
	var certPaths []string
	for _, cert := range cfg.VerificationCertificates.Certificates {
		certPaths = append(certPaths, cert.Path)
	}
	return crypto.GetSigningService("", certPaths...)
}
