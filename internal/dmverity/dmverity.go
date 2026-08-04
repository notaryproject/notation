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

// Package dmverity provides dm-verity signing functionality for OCI image layers
package dmverity

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-core-go/signature/pkcs7"
	"github.com/notaryproject/notation/v2/internal/erofs"
	"github.com/notaryproject/notation/v2/internal/registryutil"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// SignatureEnvelope holds a dm-verity PKCS#7 signature envelope for a single layer.
type SignatureEnvelope struct {
	LayerDigest string
	RootHash    string
	Signature   []byte
}

// SignatureManifest is the OCI referrer artifact containing dm-verity layer signatures.
type SignatureManifest struct {
	SchemaVersion int                  `json:"schemaVersion"`
	MediaType     string               `json:"mediaType"`
	ArtifactType  string               `json:"artifactType"`
	Config        ocispec.Descriptor   `json:"config"`
	Layers        []ocispec.Descriptor `json:"layers"`
	Subject       *ocispec.Descriptor  `json:"subject,omitempty"`
	Annotations   map[string]string    `json:"annotations,omitempty"`
}

// SignImageLayers signs all layer root hashes from OCI blob, generates PKCS#7 envelope, returns array of envelopes
func SignImageLayers(ctx context.Context, primitiveSigner signature.Signer, fetcher *registryutil.BlobFetcher, manifest ocispec.Manifest) ([]SignatureEnvelope, error) {
	var signatures []SignatureEnvelope

	for _, layer := range manifest.Layers {
		layerData, err := fetcher.FetchBlob(ctx, layer)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch layer blob %s from registry: %w", layer.Digest.String(), err)
		}

		rootHash, err := ComputeRootHash(layerData)
		if err != nil {
			return nil, fmt.Errorf("failed to generate dm-verity root hash for layer %s: %w", layer.Digest.String(), err)
		}

		sig, err := signRootHashPKCS7(primitiveSigner, rootHash)
		if err != nil {
			return nil, fmt.Errorf("failed to sign root hash for layer %s: %w", layer.Digest.String(), err)
		}

		layerSig := SignatureEnvelope{
			LayerDigest: layer.Digest.String(),
			RootHash:    rootHash,
			Signature:   sig,
		}
		signatures = append(signatures, layerSig)
	}

	return signatures, nil
}

// ComputeRootHash converts a compressed layer blob to an EROFS image and computes its root hash.
func ComputeRootHash(layerData []byte) (string, error) {
	ctx := context.Background()

	converter := erofs.NewConverter("")
	erofsData, err := converter.ConvertLayerToEROFS(ctx, layerData)
	if err != nil {
		return "", fmt.Errorf("EROFS conversion failed: %w", err)
	}

	calculator := erofs.NewVerityCalculator("")
	opts := erofs.DefaultVeritysetupOptions()
	rootHash, err := calculator.CalculateRootHash(ctx, erofsData, &opts)
	if err != nil {
		return "", fmt.Errorf("dm-verity root hash calculation failed: %w", err)
	}

	return rootHash, nil
}

// signRootHashPKCS7 creates a PKCS#7 signature envelope for the given root hash using the provided signer.
func signRootHashPKCS7(primitiveSigner signature.Signer, rootHash string) ([]byte, error) {
	env := pkcs7.NewEnvelope()

	req := &signature.SignRequest{
		Signer: primitiveSigner,
		Payload: signature.Payload{
			ContentType: pkcs7.MediaTypeEnvelope,
			Content:     []byte(rootHash),
		},
	}

	sig, err := env.Sign(req)
	if err != nil {
		return nil, fmt.Errorf("PKCS#7 signing failed: %w", err)
	}

	if len(sig) == 0 {
		return nil, fmt.Errorf("PKCS#7 signer produced empty signature for root hash %s", rootHash)
	}

	return sig, nil
}

// CreateSignatureManifest builds an OCI referrer manifest for dm-verity signatures.
func CreateSignatureManifest(signatures []SignatureEnvelope, subjectManifest ocispec.Descriptor) (*SignatureManifest, error) {
	sigManifest := &SignatureManifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		ArtifactType:  "application/vnd.cncf.notary.dmverity.v1",
		Config:        ocispec.DescriptorEmptyJSON,
		Subject:       &subjectManifest,
		Annotations: map[string]string{
			"org.opencontainers.image.created": time.Now().UTC().Format(time.RFC3339),
		},
	}

	for _, sig := range signatures {
		sigDigest := digest.FromBytes(sig.Signature)
		sigBase64 := base64.StdEncoding.EncodeToString(sig.Signature)

		layerDesc := ocispec.Descriptor{
			MediaType: "application/vnd.cncf.notary.dmverity.layer-signature+pkcs7",
			Digest:    sigDigest,
			Size:      int64(len(sig.Signature)),
			Annotations: map[string]string{
				"io.cncf.notary.dmverity.layer-digest":    sig.LayerDigest,
				"io.cncf.notary.dmverity.layer-roothash":  sig.RootHash,
				"io.cncf.notary.dmverity.layer-signature": sigBase64,
			},
		}
		sigManifest.Layers = append(sigManifest.Layers, layerDesc)
	}

	return sigManifest, nil
}
