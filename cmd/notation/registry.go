// Copyright The Notary Project Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/notaryproject/notation-go/log"
	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/cmd/notation/internal/experimental"
	notationauth "github.com/notaryproject/notation/internal/auth"
	"github.com/notaryproject/notation/internal/httputil"
	"github.com/notaryproject/notation/pkg/configutil"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
)

// inputType denotes the user input type
type inputType int

const (
	inputTypeRegistry  inputType = 1 + iota // inputType remote registry
	inputTypeOCILayout                      // inputType oci-layout
)

// getRepository returns a notationregistry.Repository given user input
// type and user input reference
func getRepository(ctx context.Context, inputType inputType, reference string, opts *SecureFlagOpts, allowReferrersAPI bool) (notationregistry.Repository, error) {
	switch inputType {
	case inputTypeRegistry:
		return getRemoteRepository(ctx, opts, reference, allowReferrersAPI)
	case inputTypeOCILayout:
		layoutPath, _, err := parseOCILayoutReference(reference)
		if err != nil {
			return nil, err
		}
		return notationregistry.NewOCIRepository(layoutPath, notationregistry.RepositoryOptions{})
	default:
		return nil, errors.New("unsupported input type")
	}
}

// getRemoteRepository returns a registry.Repository.
// When experimental feature is disabled OR allowReferrersAPI is not set,
// Notation always uses referrers tag schema to store and consume signatures
// by default.
// When experimental feature is enabled AND allowReferrersAPI is set, Notation
// tries the Referrers API, if not supported, fallback to use the Referrers
// tag schema.
//
// References:
// https://github.com/opencontainers/distribution-spec/blob/v1.1.0-rc1/spec.md#listing-referrers
// https://github.com/opencontainers/distribution-spec/blob/v1.1.0-rc1/spec.md#referrers-tag-schema
func getRemoteRepository(ctx context.Context, opts *SecureFlagOpts, reference string, allowReferrersAPI bool) (notationregistry.Repository, error) {
	logger := log.GetLogger(ctx)
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, fmt.Errorf("%q: %w. Expecting <registry>/<repository>:<tag> or <registry>/<repository>@<digest>", reference, err)
	}
	if ref.Reference == "" {
		return nil, fmt.Errorf("%q: invalid reference: no tag or digest. Expecting <registry>/<repository>:<tag> or <registry>/<repository>@<digest>", reference)
	}
	// generate notation repository
	remoteRepo, err := getRepositoryClient(ctx, opts, ref)
	if err != nil {
		return nil, err
	}

	if !experimental.IsDisabled() && allowReferrersAPI {
		logger.Info("Trying to use the referrers API")
	} else {
		logger.Info("Using the referrers tag schema")
		if err := remoteRepo.SetReferrersCapability(false); err != nil {
			return nil, err
		}
	}
	return notationregistry.NewRepository(remoteRepo), nil
}

func getRepositoryClient(ctx context.Context, opts *SecureFlagOpts, ref registry.Reference) (*remote.Repository, error) {
	authClient, insecureRegistry, err := getAuthClient(ctx, opts, ref, true)
	if err != nil {
		return nil, err
	}

	return &remote.Repository{
		Client:    authClient,
		Reference: ref,
		PlainHTTP: insecureRegistry,
	}, nil
}

func getRegistryLoginClient(ctx context.Context, opts *SecureFlagOpts, serverAddress string) (*remote.Registry, error) {
	reg, err := remote.NewRegistry(serverAddress)
	if err != nil {
		return nil, err
	}

	reg.Client, reg.PlainHTTP, err = getAuthClient(ctx, opts, reg.Reference, false)
	if err != nil {
		return nil, err
	}
	return reg, nil
}

// getAuthClient returns an *auth.Client and a bool indicating if the registry
// is insecure.
//
// If withCredential is true, the returned *auth.Client will have its Credential
// function configured.
//
// If withCredential is false, the returned *auth.Client will have a nil
// Credential function.
func getAuthClient(ctx context.Context, opts *SecureFlagOpts, ref registry.Reference, withCredential bool) (*auth.Client, bool, error) {
	var insecureRegistry bool
	if opts.InsecureRegistry {
		insecureRegistry = opts.InsecureRegistry
	} else {
		insecureRegistry = configutil.IsRegistryInsecure(ref.Registry)
		if !insecureRegistry {
			if host, _, _ := net.SplitHostPort(ref.Registry); host == "localhost" {
				insecureRegistry = true
			}
		}
	}

	// build authClient
	authClient := httputil.NewAuthClient(ctx, nil)
	if !withCredential {
		return authClient, insecureRegistry, nil
	}

	cred := opts.Credential()
	if cred != auth.EmptyCredential {
		// use the specified credential
		authClient.Credential = auth.StaticCredential(ref.Host(), cred)
	} else {
		// use saved credentials
		credsStore, err := notationauth.NewCredentialsStore()
		if err != nil {
			return nil, false, fmt.Errorf("failed to get credentials store: %w", err)
		}
		authClient.Credential = credentials.Credential(credsStore)
	}
	return authClient, insecureRegistry, nil
}
