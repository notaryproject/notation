package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/notaryproject/notation-go/log"
	notationregistry "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation/internal/trace"
	"github.com/notaryproject/notation/internal/version"
	loginauth "github.com/notaryproject/notation/pkg/auth"
	"github.com/notaryproject/notation/pkg/configutil"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

const zeroDigest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"

func getSignatureRepository(ctx context.Context, opts *SecureFlagOpts, reference string) (notationregistry.Repository, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, err
	}

	// generate notation repository
	remoteRepo, err := getRepositoryClient(ctx, opts, ref)
	if err != nil {
		return nil, err
	}
	return notationregistry.NewRepository(remoteRepo), nil
}

// getSignatureRepositoryForSign returns a registry.Repository for Sign.
// ociImageManifest denotes the type of manifest used to store signatures during
// Sign process.
// Setting ociImageManifest to true means using OCI image manifest and tag
// schema.
// Otherwise, use OCI artifact manifest and requires Referrers API.
func getSignatureRepositoryForSign(ctx context.Context, opts *SecureFlagOpts, reference string, ociImageManifest bool) (notationregistry.Repository, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, err
	}

	// generate notation repository
	remoteRepo, err := getRepositoryClient(ctx, opts, ref)
	if err != nil {
		return nil, err
	}
	// 1. OCI artifact manifest requires Referrers API.
	// Reference: https://github.com/opencontainers/distribution-spec/blob/v1.1.0-rc1/spec.md#listing-referrers
	// 2. OCI image manifest requires Referrers Tag Schema.
	// Reference: https://github.com/opencontainers/distribution-spec/blob/v1.1.0-rc1/spec.md#referrers-tag-schema
	if err := remoteRepo.SetReferrersCapability(!ociImageManifest); err != nil {
		return nil, err
	}
	// using OCI artifact manifest to store signatures. Notation requires the
	// existence of Referrers API for Sign process.
	if !ociImageManifest {
		var checkReferrerDesc ocispec.Descriptor
		checkReferrerDesc.Digest = zeroDigest
		err := remoteRepo.Referrers(ctx, checkReferrerDesc, "", func(referrers []ocispec.Descriptor) error {
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to ping Referrers API with error: %v. Try OCI image manifest using `--image-spec`", err)
		}
	}
	repositoryOpts := notationregistry.RepositoryOptions{
		OCIImageManifest: ociImageManifest,
	}
	return notationregistry.NewRepositoryWithOptions(remoteRepo, repositoryOpts), nil
}

func getRepositoryClient(ctx context.Context, opts *SecureFlagOpts, ref registry.Reference) (*remote.Repository, error) {
	authClient, plainHTTP, err := getAuthClient(ctx, opts, ref)
	if err != nil {
		return nil, err
	}

	remoteRepo := &remote.Repository{
		Client:    authClient,
		Reference: ref,
		PlainHTTP: plainHTTP,
	}
	return remoteRepo, nil
}

func getRegistryClient(ctx context.Context, opts *SecureFlagOpts, serverAddress string) (*remote.Registry, error) {
	reg, err := remote.NewRegistry(serverAddress)
	if err != nil {
		return nil, err
	}

	reg.Client, reg.PlainHTTP, err = getAuthClient(ctx, opts, reg.Reference)
	if err != nil {
		return nil, err
	}
	return reg, nil
}

func setHttpDebugLog(ctx context.Context, authClient *auth.Client) {
	if logrusLog, ok := log.GetLogger(ctx).(*logrus.Logger); ok && logrusLog.Level != logrus.DebugLevel {
		return
	}
	if authClient.Client == nil {
		authClient.Client = http.DefaultClient
	}
	if authClient.Client.Transport == nil {
		authClient.Client.Transport = http.DefaultTransport
	}
	authClient.Client.Transport = trace.NewTransport(authClient.Client.Transport)
}

func getAuthClient(ctx context.Context, opts *SecureFlagOpts, ref registry.Reference) (*auth.Client, bool, error) {
	var plainHTTP bool

	if opts.PlainHTTP {
		plainHTTP = opts.PlainHTTP
	} else {
		plainHTTP = configutil.IsRegistryInsecure(ref.Registry)
		if !plainHTTP {
			if host, _, _ := net.SplitHostPort(ref.Registry); host == "localhost" {
				plainHTTP = true
			}
		}
	}
	cred := auth.Credential{
		Username: opts.Username,
		Password: opts.Password,
	}
	if cred.Username == "" {
		cred = auth.Credential{
			RefreshToken: cred.Password,
		}
	}
	if cred == auth.EmptyCredential {
		var err error
		cred, err = getSavedCreds(ctx, ref.Registry)
		// local registry may not need credentials
		if err != nil && !errors.Is(err, loginauth.ErrCredentialsConfigNotSet) {
			return nil, false, err
		}
	}

	authClient := &auth.Client{
		Credential: func(ctx context.Context, registry string) (auth.Credential, error) {
			switch registry {
			case ref.Host():
				return cred, nil
			default:
				return auth.EmptyCredential, nil
			}
		},
		Cache:    auth.NewCache(),
		ClientID: "notation",
	}
	authClient.SetUserAgent("notation/" + version.GetVersion())

	// update authClient
	setHttpDebugLog(ctx, authClient)

	return authClient, plainHTTP, nil
}

func getSavedCreds(ctx context.Context, serverAddress string) (auth.Credential, error) {
	nativeStore, err := loginauth.GetCredentialsStore(ctx, serverAddress)
	if err != nil {
		return auth.EmptyCredential, err
	}

	return nativeStore.Get(serverAddress)
}
