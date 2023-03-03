package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/notaryproject/notation-go/log"
	notationregistry "github.com/notaryproject/notation-go/registry"
	notationerrors "github.com/notaryproject/notation/cmd/notation/internal/errors"
	"github.com/notaryproject/notation/internal/trace"
	"github.com/notaryproject/notation/internal/version"
	loginauth "github.com/notaryproject/notation/pkg/auth"
	"github.com/notaryproject/notation/pkg/configutil"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/errcode"
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
// Setting ociImageManifest to true means using OCI image manifest and the
// Referrers tag schema.
// Otherwise, use OCI artifact manifest and requires the Referrers API.
func getSignatureRepositoryForSign(ctx context.Context, opts *SecureFlagOpts, reference string, ociImageManifest bool) (notationregistry.Repository, error) {
	logger := log.GetLogger(ctx)
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, err
	}

	// generate notation repository
	remoteRepo, err := getRepositoryClient(ctx, opts, ref)
	if err != nil {
		return nil, err
	}

	// Notation enforces the following two paths during Sign process:
	// 1. OCI artifact manifest uses the Referrers API
	// Reference: https://github.com/opencontainers/distribution-spec/blob/v1.1.0-rc1/spec.md#listing-referrers
	// 2. OCI image manifest uses the Referrers Tag Schema
	// Reference: https://github.com/opencontainers/distribution-spec/blob/v1.1.0-rc1/spec.md#referrers-tag-schema
	if !ociImageManifest {
		logger.Info("Use OCI artifact manifest and Referrers API to store signature")
		// ping Referrers API
		if err := pingReferrersAPI(ctx, remoteRepo); err != nil {
			return nil, err
		}
		logger.Info("Successfully pinged Referrers API on target registry")
	} else {
		logger.Info("Use OCI image manifest and Referrers Tag Schema to store signature")
		if err := remoteRepo.SetReferrersCapability(false); err != nil {
			return nil, err
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

	return &remote.Repository{
		Client:    authClient,
		Reference: ref,
		PlainHTTP: plainHTTP,
	}, nil
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

func pingReferrersAPI(ctx context.Context, remoteRepo *remote.Repository) error {
	logger := log.GetLogger(ctx)
	if err := remoteRepo.SetReferrersCapability(true); err != nil {
		return err
	}
	var checkReferrerDesc ocispec.Descriptor
	checkReferrerDesc.Digest = zeroDigest
	// core process
	err := remoteRepo.Referrers(ctx, checkReferrerDesc, "", func(referrers []ocispec.Descriptor) error {
		return nil
	})
	if err != nil {
		var errResp *errcode.ErrorResponse
		if !errors.As(err, &errResp) || errResp.StatusCode != http.StatusNotFound {
			return err
		}
		if isErrorCode(errResp, errcode.ErrorCodeNameUnknown) {
			// The repository is not found in the target registry.
			// This is triggered when putting signatures to an empty repository.
			// For notation, this path should never be triggered.
			return err
		}
		// A 404 returned by Referrers API indicates that Referrers API is
		// not supported.
		logger.Infof("failed to ping Referrers API with error: %v", err)
		errMsg := "Target registry does not support the Referrers API. Try removing the flag `--signature-manifest artifact` to store signatures using OCI image manifest"
		return notationerrors.ErrorReferrersAPINotSupported{Msg: errMsg}
	}
	return nil
}

// isErrorCode returns true if err is an Error and its Code equals to code.
func isErrorCode(err error, code string) bool {
	var ec errcode.Error
	return errors.As(err, &ec) && ec.Code == code
}

// ociLayoutFolderAsRepositoryForSign returns a oci.Store as registry.Repository
func ociLayoutFolderAsRepositoryForSign(path string, ociImageManifest bool) (notationregistry.Repository, error) {
	root, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute representation of path: %w", err)
	}
	ociStore, err := oci.New(root)
	if err != nil {
		return nil, fmt.Errorf("failed to create OCI store: %w", err)
	}
	repositoryOpts := notationregistry.RepositoryOptions{
		OCIImageManifest: ociImageManifest,
	}
	return notationregistry.NewRepositoryWithOptions(ociStore, repositoryOpts), nil
}

// ociLayoutFolderAsRepository returns a oci.Store as registry.Repository
func ociLayoutFolderAsRepository(path string) (notationregistry.Repository, error) {
	root, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute representation of path: %w", err)
	}
	ociStore, err := oci.New(root)
	if err != nil {
		return nil, fmt.Errorf("failed to create OCI store: %w", err)
	}
	return notationregistry.NewRepository(ociStore), nil
}

// ociLayoutTarForSign returns a oci.ReadOnlyStore as registry.Repository
func ociLayoutTarForSign(path string, ociImageManifest bool) (notationregistry.Repository, error) {
	root, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute representation of path: %w", err)
	}
	_, err = oci.NewStorageFromTar(root)
	if err != nil {
		return nil, notationerrors.ErrorOciLayoutTarForSign{Msg: err.Error()}
	}
	return nil, nil
}

// ociLayoutTar returns a oci.ReadOnlyStore as registry.Repository
// func ociLayoutTar(path string) (notationregistry.Repository, error) {
// 	root, err := filepath.Abs(path)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get absolute representation of path: %w", err)
// 	}
// 	_, err = oci.NewStorageFromTar(root)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create ReadOnlyStorage from tar: %w", err)
// 	}
// 	return nil, nil
// }

// parseOCILayoutReference parses the raw in format of <path>[:<tag>|@<digest>]
func parseOCILayoutReference(raw string) (path string, ref string, err error) {
	if idx := strings.LastIndex(raw, "@"); idx != -1 {
		// `digest` found
		path = raw[:idx]
		ref = raw[idx+1:]
	} else {
		// find `tag`
		i := strings.LastIndex(raw, ":")
		if i < 0 || (i == 1 && len(raw) > 2 && unicode.IsLetter(rune(raw[0])) && raw[2] == '\\') {
			return "", "", errors.New("reference is missing digest or tag")
		} else {
			path, ref = raw[:i], raw[i+1:]
		}
		if path == "" {
			return "", "", fmt.Errorf("found empty file path in %q", raw)
		}
	}
	return
}
