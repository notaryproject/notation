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

package osutil

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// WriteFile writes to a path with all parent directories created.
func WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// WriteFileWithPermission writes to a path with all parent directories created.
func WriteFileWithPermission(path string, data []byte, perm fs.FileMode, overwrite bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	flag := os.O_WRONLY | os.O_CREATE
	if overwrite {
		flag |= os.O_TRUNC
	} else {
		flag |= os.O_EXCL
	}
	file, err := os.OpenFile(path, flag, perm)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

// CopyToDir copies the src file to dst. Existing file will be overwritten.
func CopyToDir(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	if err := os.MkdirAll(dst, 0700); err != nil {
		return 0, err
	}
	dstFile := filepath.Join(dst, filepath.Base(src))
	destination, err := os.Create(dstFile)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	err = destination.Chmod(0600)
	if err != nil {
		return 0, err
	}
	return io.Copy(destination, source)
}

// IsRegularFile checks if path is a regular file
func IsRegularFile(path string) (bool, error) {
	fileStat, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileStat.Mode().IsRegular(), nil
}

// DetectFileType returns a file's content type given path
func DetectFileType(path string) (string, error) {
	rc, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer rc.Close()
	lr := io.LimitReader(rc, 512)
	header := make([]byte, 512)
	if _, err := lr.Read(header); err != nil {
		return "", err
	}
	return http.DetectContentType(header), nil
}

// FileNameWithoutExtension returns the file name without extension.
// For example,
// when input is xyz.exe, output is xyz
// when input is xyz.tar.gz, output is xyz.tar
func FileNameWithoutExtension(inputName string) string {
	fileName := filepath.Base(inputName)
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

// ValidateChecksum returns nil if SHA256 of file at path equals to checksum.
func ValidateChecksum(path string, checksum string) error {
	rc, err := os.Open(path)
	if err != nil {
		return err
	}
	defer rc.Close()
	sha256Hash := sha256.New()
	if _, err := io.Copy(sha256Hash, rc); err != nil {
		return err
	}
	sha256sum := sha256Hash.Sum(nil)
	enc := strings.ToLower(hex.EncodeToString(sha256sum[:]))
	if enc != strings.ToLower(checksum) {
		return fmt.Errorf("plugin checksum does not match user input. Expecting %s", checksum)
	}
	return nil
}
