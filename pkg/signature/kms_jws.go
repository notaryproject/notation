package signature

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/notaryproject/notation-go-lib"
	"github.com/notaryproject/notation/pkg/config"
	"github.com/notaryproject/notation/pkg/executor"
)

// SignRequest is the request to sign artifacts
type SignRequest struct {
	Version     string                 `json:"version"`
	Descriptor  notation.Descriptor    `json:"descriptor"`
	SignOptions notation.SignOptions   `json:"signOptions"`
	KMSProfile  config.KMSProfileSuite `json:"kmsProfile"`
}

// VerifyRequest is the request to verify a signature.
type VerifyRequest struct {
	Version       string                 `json:"version"`
	Signature     []byte                 `json:"signature"`
	VerifyOptions notation.VerifyOptions `json:"verifyOptions"`
	KMSProfile    config.KMSProfileSuite `json:"kmsProfile"`
}

type externalPlugin struct {
	kmsProfile config.KMSProfileSuite

	executor executor.Executor
}

// NewSignerWithPlugin returns a signer that uses the given plugin
func NewSignerWithPlugin(kmsProfile config.KMSProfileSuite, pluginPath string) (notation.Signer, error) {
	if pluginPath == "" {
		return nil, errors.New("plugin path not specified")
	}

	// create signer
	return &externalPlugin{
		kmsProfile: kmsProfile,
		executor:   executor.NewExecutor(pluginPath, "sign"),
	}, nil
}

func NewVerifierWithPlugin(kmsProfile config.KMSProfileSuite, pluginPath string) (notation.Verifier, error) {
	if pluginPath == "" {
		return nil, errors.New("plugin path not specified")
	}

	// create verifier
	return &externalPlugin{
		kmsProfile: kmsProfile,
		executor:   executor.NewExecutor(pluginPath, "verify"),
	}, nil
}

func (p *externalPlugin) Sign(ctx context.Context, desc notation.Descriptor, opts notation.SignOptions) ([]byte, error) {
	// create request
	req := SignRequest{
		Version:     "v0.1.0-alpha.0",
		Descriptor:  desc,
		SignOptions: opts,
		KMSProfile:  p.kmsProfile,
	}

	// marshal request
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// execute plugin
	return p.executor.Execute(ctx, "%s", string(reqBytes))
}

func (p *externalPlugin) Verify(ctx context.Context, signature []byte, opts notation.VerifyOptions) (notation.Descriptor, error) {
	retDesc := notation.Descriptor{}

	// create request
	req := VerifyRequest{
		Version:       "v0.1.0-alpha.0",
		Signature:     signature,
		VerifyOptions: opts,
		KMSProfile:    p.kmsProfile,
	}

	// marshal request
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return retDesc, err
	}

	// execute plugin
	respBytes, err := p.executor.Execute(ctx, string(reqBytes))
	if err != nil {
		return retDesc, err
	}

	// unmarshal response
	err = json.Unmarshal(respBytes, &retDesc)
	if err != nil {
		return retDesc, err
	}

	return retDesc, nil
}
