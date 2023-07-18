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

package cert

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCertGenerateCommand(t *testing.T) {
	opts := &certGenerateTestOpts{}
	cmd := certGenerateTestCommand(opts)
	expected := &certGenerateTestOpts{
		name:      "name",
		bits:      2048,
		isDefault: true,
	}
	if err := cmd.ParseFlags([]string{
		"name",
		"--bits", fmt.Sprint(expected.bits),
		"--default"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect cert generate-test opts: %v, got: %v", expected, opts)
	}
}

func TestCertGenerateTestCommand_MissingArgs(t *testing.T) {
	cmd := certGenerateTestCommand(nil)
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}
