package main

import (
	"os"
	"os/exec"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/notaryproject/notation/pkg/signature"
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

func getVerifier() (notation.Verifier, error) {
	cfg, err := config.LoadOrDefaultOnce()
	if err != nil {
		return nil, err
	}
	var certPaths []string
	for _, cert := range cfg.VerificationCertificates.Certificates {
		certPaths = append(certPaths, cert.Path)
	}
	return signature.NewVerifierFromFiles(certPaths)
}
