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
	"strings"
	"testing"
)

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
