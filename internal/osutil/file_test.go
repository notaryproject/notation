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
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func validFileContent(t *testing.T, filename string, content []byte) {
	b, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(content, b) {
		t.Fatal("file content is not correct")
	}
}

func TestWriteFile(t *testing.T) {
	t.Run("write file", func(t *testing.T) {
		tempDir := t.TempDir()
		data := []byte("data")
		filename := filepath.Join(tempDir, "a", "file.txt")
		if err := WriteFile(filename, data); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("write file with directory permission error", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping test on Windows")
		}

		tempDir := t.TempDir()
		data := []byte("data")
		// forbid writing to tempDir
		if err := os.Chmod(tempDir, 0000); err != nil {
			t.Fatal(err)
		}
		filename := filepath.Join(tempDir, "a", "file.txt")
		if err := WriteFile(filename, data); err == nil {
			t.Fatal("should write failed")
		}
	})

	t.Run("check file correctness", func(t *testing.T) {
		tempDir := t.TempDir()
		data := []byte("data")
		filename := filepath.Join(tempDir, "a", "file.txt")
		if err := WriteFile(filename, data); err != nil {
			t.Fatal(err)
		}
		validFileContent(t, filename, data)
	})
}

func TestWriteFileWithPermission(t *testing.T) {
	t.Run("write without override", func(t *testing.T) {
		tempDir := t.TempDir()
		data := []byte("data")
		filename := filepath.Join(tempDir, "file.txt")
		if err := WriteFileWithPermission(filename, data, 0644, false); err != nil {
			t.Fatal(err)
		}

		if err := WriteFileWithPermission(filename, data, 0644, false); err == nil {
			t.Fatal("should have an error")
		}
	})

	t.Run("write with override", func(t *testing.T) {
		tempDir := t.TempDir()
		data := []byte("data")
		filename := filepath.Join(tempDir, "file.txt")
		if err := WriteFileWithPermission(filename, data, 0644, false); err != nil {
			t.Fatal(err)
		}

		if err := WriteFileWithPermission(filename, data, 0644, true); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("write with directory permission error", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping test on Windows")
		}

		tempDir := t.TempDir()
		data := []byte("data")
		filename := filepath.Join(tempDir, "a", "file.txt")
		// forbid writing to tempDir
		if err := os.Chmod(tempDir, 0000); err != nil {
			t.Fatal(err)
		}
		if err := WriteFileWithPermission(filename, data, 0644, false); err == nil {
			t.Fatal("should have an error")
		}
	})

	t.Run("valid file content", func(t *testing.T) {
		tempDir := t.TempDir()
		data := []byte("data")
		filename := filepath.Join(tempDir, "file.txt")
		if err := WriteFileWithPermission(filename, data, 0644, false); err != nil {
			t.Fatal(err)
		}

		if err := WriteFileWithPermission(filename, data, 0644, false); err == nil {
			t.Fatal("should have an error")
		}

		validFileContent(t, filename, data)
	})
}

func TestCopyToDir(t *testing.T) {
	t.Run("copy file", func(t *testing.T) {
		tempDir := t.TempDir()
		data := []byte("data")
		filename := filepath.Join(tempDir, "a", "file.txt")
		if err := WriteFile(filename, data); err != nil {
			t.Fatal(err)
		}

		destDir := filepath.Join(tempDir, "b")
		if _, err := CopyToDir(filename, destDir); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("source directory permission error", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping test on Windows")
		}

		tempDir := t.TempDir()
		destDir := t.TempDir()
		data := []byte("data")
		filename := filepath.Join(tempDir, "a", "file.txt")
		if err := WriteFile(filename, data); err != nil {
			t.Fatal(err)
		}

		if err := os.Chmod(tempDir, 0000); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(tempDir, 0700)

		if _, err := CopyToDir(filename, destDir); err == nil {
			t.Fatal("should have error")
		}
	})

	t.Run("not a regular file", func(t *testing.T) {
		tempDir := t.TempDir()
		destDir := t.TempDir()
		if _, err := CopyToDir(tempDir, destDir); err == nil {
			t.Fatal("should have error")
		}
	})

	t.Run("source file permission error", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping test on Windows")
		}

		tempDir := t.TempDir()
		destDir := t.TempDir()
		data := []byte("data")
		// prepare file
		filename := filepath.Join(tempDir, "a", "file.txt")
		if err := WriteFile(filename, data); err != nil {
			t.Fatal(err)
		}
		// forbid reading
		if err := os.Chmod(filename, 0000); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(filename, 0600)
		if _, err := CopyToDir(filename, destDir); err == nil {
			t.Fatal("should have error")
		}
	})

	t.Run("dest directory permission error", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping test on Windows")
		}

		tempDir := t.TempDir()
		destTempDir := t.TempDir()
		data := []byte("data")
		// prepare file
		filename := filepath.Join(tempDir, "a", "file.txt")
		if err := WriteFile(filename, data); err != nil {
			t.Fatal(err)
		}
		// forbid dest directory operation
		if err := os.Chmod(destTempDir, 0000); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(destTempDir, 0700)
		if _, err := CopyToDir(filename, filepath.Join(destTempDir, "a")); err == nil {
			t.Fatal("should have error")
		}
	})

	t.Run("dest directory permission error 2", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping test on Windows")
		}

		tempDir := t.TempDir()
		destTempDir := t.TempDir()
		data := []byte("data")
		// prepare file
		filename := filepath.Join(tempDir, "a", "file.txt")
		if err := WriteFile(filename, data); err != nil {
			t.Fatal(err)
		}
		// forbid writing to destTempDir
		if err := os.Chmod(destTempDir, 0000); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(destTempDir, 0700)
		if _, err := CopyToDir(filename, destTempDir); err == nil {
			t.Fatal("should have error")
		}
	})

	t.Run("copy file and check content", func(t *testing.T) {
		tempDir := t.TempDir()
		data := []byte("data")
		filename := filepath.Join(tempDir, "a", "file.txt")
		if err := WriteFile(filename, data); err != nil {
			t.Fatal(err)
		}

		destDir := filepath.Join(tempDir, "b")
		if _, err := CopyToDir(filename, destDir); err != nil {
			t.Fatal(err)
		}
		validFileContent(t, filepath.Join(destDir, "file.txt"), data)
	})
}

func TestValidateChecksum(t *testing.T) {
	expectedErrorMsg := "plugin SHA-256 checksum does not match user input. Expecting abcd123"
	if err := ValidateSHA256Sum("./testdata/test", "abcd123"); err == nil || err.Error() != expectedErrorMsg {
		t.Fatalf("expected err %s, but got %v", expectedErrorMsg, err)
	}
	if err := ValidateSHA256Sum("./testdata/test", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"); err != nil {
		t.Fatalf("expected nil err, but got %v", err)
	}
}
