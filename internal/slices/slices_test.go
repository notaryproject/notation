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

package slices

import (
	"testing"
)

func TestContainerElement(t *testing.T) {
	tests := []struct {
		c    []string
		v    string
		want bool
	}{
		{nil, "", false},
		{[]string{}, "", false},
		{[]string{"1", "2", "3"}, "4", false},
		{[]string{"1", "2", "3"}, "2", true},
		{[]string{"1", "2", "2", "3"}, "2", true},
		{[]string{"1", "2", "3", "2"}, "2", true},
	}
	for _, tt := range tests {
		if got := Contains(tt.c, tt.v); got != tt.want {
			t.Errorf("ContainerElement() = %v, want %v", got, tt.want)
		}
	}
}
