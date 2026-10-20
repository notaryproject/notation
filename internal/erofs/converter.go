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

// Package erofs converts OCI layer tarballs to EROFS filesystem format
// using mkfs.erofs --tar=i (tar index mode). The output must match
// containerd's erofs-snapshotter for compatible dm-verity root hashes.
package erofs

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

const (
	mkfsErofsTimeout  = 5 * time.Minute
	blockAlignment    = 512                                    // Must match erofs-snapshotter EROFS_BLOCK_ALIGNMENT
	fixedMetadataUUID = "c1b9d5a2-f162-11cf-9ece-0020afc76f16" // Must match erofs-snapshotter EROFS_METADATA_UUID
)

// Converter converts OCI layers to EROFS using tar-index mode.
type Converter struct {
	TempDir string
}

// NewConverter creates a new EROFS converter. If tempDir is empty, os.TempDir() is used.
func NewConverter(tempDir string) *Converter {
	if tempDir == "" {
		tempDir = os.TempDir()
	}
	return &Converter{
		TempDir: tempDir,
	}
}

// ConvertLayerToEROFS converts a compressed OCI layer (tar.gz) to EROFS format.
// Returns EROFS metadata + tar data, aligned to 512-byte boundary for dm-verity.
func (c *Converter) ConvertLayerToEROFS(ctx context.Context, layerData []byte) ([]byte, error) {
	if len(layerData) == 0 {
		return nil, fmt.Errorf("layer data is empty")
	}

	tarData, err := c.decompressGzip(layerData)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress gzip: %w", err)
	}

	erofsData, err := c.buildEROFSImage(ctx, tarData)
	if err != nil {
		return nil, fmt.Errorf("failed to build EROFS image: %w", err)
	}

	return erofsData, nil
}

// decompressGzip decompresses gzip-compressed data and returns the raw tar data.
func (c *Converter) decompressGzip(compressedData []byte) ([]byte, error) {
	gzReader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("layer data is not valid gzip (expected tar.gz layer): %w", err)
	}
	defer gzReader.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, gzReader); err != nil {
		return nil, fmt.Errorf("gzip decompression failed (corrupted layer data?): %w", err)
	}
	return buf.Bytes(), nil
}

// buildEROFSImage creates an EROFS image from tar data using mkfs.erofs --tar=i.
// It generates EROFS metadata, appends the original tar data, and aligns the result
// to 512 bytes for dm-verity compatibility.
func (c *Converter) buildEROFSImage(ctx context.Context, tarData []byte) ([]byte, error) {
	if _, err := exec.LookPath("mkfs.erofs"); err != nil {
		return nil, fmt.Errorf("mkfs.erofs not found in PATH: install 'erofs-utils' package (apt install erofs-utils / dnf install erofs-utils): %w", err)
	}

	tarFile, err := os.CreateTemp(c.TempDir, "layer-*.tar")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp tar file: %w", err)
	}
	tarPath := tarFile.Name()
	defer os.Remove(tarPath)

	if _, err := tarFile.Write(tarData); err != nil {
		tarFile.Close()
		return nil, fmt.Errorf("failed to write tar data: %w", err)
	}
	tarFile.Close()

	erofsFile, err := os.CreateTemp(c.TempDir, "erofs-metadata-*.img")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp EROFS file: %w", err)
	}
	erofsPath := erofsFile.Name()
	erofsFile.Close()
	defer os.Remove(erofsPath)

	cmdCtx, cancel := context.WithTimeout(ctx, mkfsErofsTimeout)
	defer cancel()

	// Flags must match erofs-snapshotter's mkfs.erofs invocation
	cmd := exec.CommandContext(cmdCtx, "mkfs.erofs",
		"--tar=i", // tar index mode
		"-T", "0", // Zero unix time
		"--mkfs-time",           // Clear mkfs time in superblock
		"-U", fixedMetadataUUID, // Fixed UUID for deterministic builds
		"--aufs",  // Convert OCI whiteouts to overlayfs metadata
		"--quiet", // Quiet mode
		erofsPath,
		tarPath,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("mkfs.erofs failed: %w, stdout: %s, stderr: %s", err, stdout.String(), stderr.String())
	}

	erofsMetadata, err := os.ReadFile(erofsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read EROFS metadata: %w", err)
	}

	combinedData := appendTarData(erofsMetadata, tarData)
	alignedData := alignTo512(combinedData)

	return alignedData, nil
}

func appendTarData(erofsMetadata, tarData []byte) []byte {
	combined := make([]byte, len(erofsMetadata)+len(tarData))
	copy(combined, erofsMetadata)
	copy(combined[len(erofsMetadata):], tarData)
	return combined
}

// alignTo512 pads data to 512-byte boundary for dm-verity.
func alignTo512(data []byte) []byte {
	remainder := len(data) % blockAlignment
	if remainder == 0 {
		return data
	}

	paddingSize := blockAlignment - remainder
	aligned := make([]byte, len(data)+paddingSize)
	copy(aligned, data)
	return aligned
}
