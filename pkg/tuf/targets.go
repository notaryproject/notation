package tuf

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/docker/go/canonical/json"
	"github.com/notaryproject/nv2/pkg/reference"
	"github.com/theupdateframework/notary/tuf/data"
	"github.com/theupdateframework/notary/tuf/signed"
)

// Target represents a TUF target with name
type Target struct {
	data.FileMeta
	Name string `json:"name"`
}

// TargetMetadata describes the target content with extra information
type TargetMetadata struct {
	AccessedAt time.Time `json:"accessedAt,omitempty"`
	MediaType  string    `json:"mediaType,omitempty"`
}

// NewTarget is a helper method that returns a Target
func NewTarget(manifest *reference.Manifest) (*Target, error) {
	metadata := TargetMetadata{
		AccessedAt: manifest.AccessedAt,
		MediaType:  manifest.MediaType,
	}
	metadataJSON, err := json.MarshalCanonical(metadata)
	if err != nil {
		return nil, err
	}
	tartgetCustom := new(json.RawMessage)
	if err := tartgetCustom.UnmarshalJSON(metadataJSON); err != nil {
		return nil, err
	}

	hashes := make(data.Hashes)
	for _, digest := range manifest.Digests {
		alg := digest.Algorithm().String()
		hash, err := hex.DecodeString(digest.Encoded())
		if err != nil {
			return nil, err
		}
		hashes[alg] = hash
	}

	return &Target{
		FileMeta: data.FileMeta{
			Hashes: hashes,
			Length: manifest.Size,
			Custom: tartgetCustom,
		},
		Name: manifest.Name,
	}, nil
}

// AddTargets adds targets to the existing targets.
func AddTargets(base *data.Targets, targets ...*Target) *data.Targets {
	if base == nil {
		base = &data.NewTargets().Signed
	}

	for _, target := range targets {
		base.Targets[target.Name] = target.FileMeta
	}
	base.Expires = time.Now().UTC().Add(DefaultTargetExpiry)
	base.Version++
	return base
}

// SignTargets signs the targets
func SignTargets(ctx context.Context, signer Signer, targets *data.Targets) (*Signed, error) {
	signedTargets := data.SignedTargets{
		Signed: *targets,
	}
	tufSigned, err := signedTargets.ToSigned()
	if err != nil {
		return nil, err
	}
	signed := SignedFromTUF(tufSigned)
	err = Sign(ctx, signer, signed)
	if err != nil {
		return nil, err
	}
	return signed, nil
}

// VerifyTargets verifies the targets
func VerifyTargets(ctx context.Context, verifier Verifier, signedContent *Signed, minVersion int) (*data.Targets, error) {
	_, err := Verify(ctx, verifier, signedContent)
	if err != nil {
		return nil, err
	}

	signedTargets, err := data.TargetsFromSigned(signedContent.ToTUF(), data.CanonicalTargetsRole)
	if err != nil {
		return nil, err
	}
	targets := signedTargets.Signed

	if err := signed.VerifyExpiry(&targets.SignedCommon, data.CanonicalTargetsRole); err != nil {
		return nil, err
	}
	if err := signed.VerifyVersion(&targets.SignedCommon, minVersion); err != nil {
		return nil, err
	}
	return &targets, nil
}

// IsManifestInTargets checks if a manifest is referenced by the targets
func IsManifestInTargets(manifest *reference.Manifest, targets *data.Targets) bool {
	if name := manifest.Name; name != "" {
		target, ok := targets.Targets[name]
		return ok && ManifestMatchesTarget(manifest, &target)
	}

	for _, target := range targets.Targets {
		if found := ManifestMatchesTarget(manifest, &target); found {
			return true
		}
	}
	return false
}

// ManifestMatchesTarget checks if a manifest is referenced by the specified target
func ManifestMatchesTarget(manifest *reference.Manifest, target *data.FileMeta) bool {
	if manifest.Size != target.Length {
		return false
	}

	found := false
	for _, digest := range manifest.Digests {
		alg := digest.Algorithm().String()
		hash, ok := target.Hashes[alg]
		if !ok {
			continue
		}
		if hex.EncodeToString(hash) == digest.Encoded() {
			found = true
			break
		}
	}
	if !found {
		return false
	}

	if target.Custom == nil {
		return false
	}
	var metadata TargetMetadata
	if err := json.Unmarshal(*target.Custom, &metadata); err != nil {
		return false
	}
	return metadata.MediaType == manifest.MediaType
}
