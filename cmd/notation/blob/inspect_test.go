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

package blob

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestBlobInspectCommand_MissingArgs(t *testing.T) {
	command := inspectCommand(nil)
	if err := command.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}

func TestReadFile(t *testing.T) {
	noFile := filepath.FromSlash("")
	expectedErr := errors.New("open : no such file or directory")
	_, err := readFile(noFile)
	if err == nil || err.Error() != "open : no such file or directory" {
		t.Fatalf("expected err: %v, got: %v", expectedErr, err)
	}

	emptyFile := filepath.FromSlash("../../../internal/testdata/Empty.txt")
	expectedErr = errors.New("file is empty")
	_, err = readFile(emptyFile)
	if err == nil || err.Error() != "file is empty" {
		t.Fatalf("expected err: %v, got: %v", expectedErr, err)
	}

	filePath := filepath.FromSlash("../../../internal/testdata/Output.txt")
	expectedErr = errors.New("unable to read as file size was greater than 10485760 bytes")
	_, err = readFile(filePath)
	if err == nil || err.Error() != "unable to read as file size was greater than 10485760 bytes" {
		t.Fatalf("expected err: %v, got: %v", expectedErr, err)
	}
}
