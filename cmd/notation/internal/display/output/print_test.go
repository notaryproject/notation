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

// copied and adopted from https://github.com/oras-project/oras with
// modification
/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package output

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

type mockWriter struct {
	errorCount int
	written    string
}

func (mw *mockWriter) Write(p []byte) (n int, err error) {
	mw.written += string(p)
	if strings.TrimSpace(string(p)) != "boom" {
		return len(string(p)), nil
	}
	mw.errorCount++
	return 0, fmt.Errorf("boom %s", string(p))
}

func (mw *mockWriter) String() string {
	return mw.written
}

func TestPrinter_Print(t *testing.T) {
	mockWriter := &mockWriter{}
	printer := NewPrinter(mockWriter, os.Stderr)

	t.Run("Println success", func(t *testing.T) {
		err := printer.Println("hello")
		if err != nil {
			t.Errorf("Expected no error got <%v>", err)
		}
		if mockWriter.String() != "hello\n" {
			t.Errorf("Expected hello got <%s>", mockWriter.String())
		}
	})
	t.Run("Println failed", func(t *testing.T) {
		err := printer.Println("boom")
		if mockWriter.errorCount != 1 {
			t.Errorf("Expected one error actual <%d>", mockWriter.errorCount)
		}
		if err == nil {
			t.Error("Expected error got <nil>")
		}
	})
	t.Run("Printf failed", func(t *testing.T) {
		err := printer.Printf("boom")
		if mockWriter.errorCount != 2 {
			t.Errorf("Expected two errors actual <%d>", mockWriter.errorCount)
		}
		if err == nil {
			t.Error("Expected error got <nil>")
		}
	})
}

func Test_PrintPrettyJSON(t *testing.T) {
	builder := &strings.Builder{}
	given := map[string]int{"bob": 5}
	expected := "{\n  \"bob\": 5\n}\n"
	err := PrintPrettyJSON(builder, given)
	if err != nil {
		t.Error("Expected no error got <" + err.Error() + ">")
	}
	actual := builder.String()
	if expected != actual {
		t.Error("Expected <" + expected + "> not equal to actual <" + actual + ">")
	}
}
