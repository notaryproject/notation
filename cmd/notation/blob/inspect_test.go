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
	"github.com/notaryproject/notation/cmd/notation/internal/osutil"
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
	_, err := osutil.ReadFile(noFile)
	if err == nil || err.Error() != "open : no such file or directory" {
		t.Fatalf("expected err: %v, got: %v", expectedErr, err)
	}

	emptyFile := filepath.FromSlash("../../../internal/testdata/Empty.txt")
	expectedErr = errors.New("file is empty")
	_, err = osutil.ReadFile(emptyFile)
	if err == nil || err.Error() != "file is empty" {
		t.Fatalf("expected err: %v, got: %v", expectedErr, err)
	}

	largeFile := filepath.FromSlash("../../../internal/testdata/LargeFile.txt")
	expectedErr = errors.New("unable to read as file size was greater than 10485760 bytes")
	_, err = osutil.ReadFile(largeFile)
	if err == nil || err.Error() != "unable to read as file size was greater than 10485760 bytes" {
		t.Fatalf("expected err: %v, got: %v", expectedErr, err)
	}

	file := filepath.FromSlash("../../../internal/testdata/File.txt")
	contents, err := osutil.ReadFile(file)
	if err != nil {
		t.Fatalf("Reading file failed: %v", err)
	}
	if string(contents) != "awesome notation\n" {
		t.Fatalf("Reading contents of file failed")
	}
}
