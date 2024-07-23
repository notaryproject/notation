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
	"fmt"
	"github.com/notaryproject/notation-go/dir"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestReadFile(t *testing.T) {
	t.Run("no file", func(t *testing.T) {
		tempRoot := t.TempDir()
		dir.UserConfigDir = tempRoot
		path := ""
		expectedErrMsg := "open : no such file or directory"
		_, err := ReadFile(path, 0)
		if err == nil || err.Error() != "open : no such file or directory" {
			t.Fatalf("expected err: %v, got: %v", expectedErrMsg, err)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		tempRoot := t.TempDir()
		dir.UserConfigDir = tempRoot
		path := filepath.Join(tempRoot, "emptyFile.txt")
		if err := os.WriteFile(path, []byte(""), 0600); err != nil {
			t.Fatalf("TestReadFile create empty file failed. Error: %v", err)
		}
		expectedErrMsg := "file is empty"
		_, err := ReadFile(path, 0)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected err: %s, got: %v", expectedErrMsg, err)
		}
	})

	t.Run("correct file", func(t *testing.T) {
		tempRoot := t.TempDir()
		dir.UserConfigDir = tempRoot
		path := filepath.Join(tempRoot, "correctFile.txt")
		if err := os.WriteFile(path, []byte("awesome notation\n"), 0600); err != nil {
			t.Fatalf("TestReadFile create correct file failed. Error: %v", err)
		}
		contents, err := ReadFile(path, 20)
		if err != nil {
			t.Fatalf("Reading file failed: %v", err)
		}
		if string(contents) != "awesome notation\n" {
			t.Fatalf("Reading contents of file failed")
		}
	})

	// Test for file size larger than expected
	t.Run("large file", func(t *testing.T) {
		tempRoot := t.TempDir()
		dir.UserConfigDir = tempRoot
		path := filepath.Join(tempRoot, "largeFile.txt")
		file, err := os.Create(path)
		if err != nil {
			log.Fatal("Failed to create output")
		}
		fileSize := int64(2 * 1024 * 1024) //2 Mb in bytes
		_, err = file.Seek(fileSize-1, 0)
		if err != nil {
			log.Fatal("Failed to seek")
		}
		defer file.Close()
		_, err = file.Write([]byte{0})
		if err != nil {
			log.Fatal("Write failed")
		}
		err = file.Close()
		if err != nil {
			log.Fatal("Failed to close file")
		}
		expectedSize := int64(1048576) //1 Mb in bytes
		expectedErrMsg := fmt.Sprintf("unable to read as file size is greater than %v bytes", expectedSize)
		_, err = ReadFile(path, expectedSize)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected err: %s, got: %v", expectedErrMsg, err)
		}
	})
}
