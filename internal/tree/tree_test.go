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

package tree

import (
	"reflect"
	"testing"
)

func TestNodeCreation(t *testing.T) {
	node := New("root")
	expected := Node{Value: "root"}

	if !reflect.DeepEqual(*node, expected) {
		t.Fatalf("expected %+v, got %+v", expected, *node)
	}
}

func TestNodeAdd(t *testing.T) {
	root := New("root")
	root.Add("child")

	if !root.ContainsChild("child") {
		t.Error("expected root to have child node with value 'child'")
		t.Fatalf("actual root: %+v", root)
	}
}

func TestNodeAddPair(t *testing.T) {
	root := New("root")
	root.AddPair("key", "value")

	if !root.ContainsChild("key: value") {
		t.Error("expected root to have child node with value 'key: value'")
		t.Fatalf("actual root: %+v", root)
	}
}

func ExampleRootPrint() {
	root := New("root")
	root.Print()

	// Output:
	// root
}

func ExampleSingleLayerPrint() {
	root := New("root")
	root.Add("child1")
	root.Add("child2")
	root.Print()

	// Output:
	// root
	// ├── child1
	// └── child2
}

func ExampleMultiLayerPrint() {
	root := New("root")
	child1 := root.Add("child1")
	child1.AddPair("key", "value")
	child2 := root.Add("child2")
	child2.Add("child2.1")
	child2.Add("child2.2")
	root.Print()

	// Output:
	// root
	// ├── child1
	// │   └── key: value
	// └── child2
	//     ├── child2.1
	//     └── child2.2
}

func (n *Node) ContainsChild(value string) bool {
	for _, child := range n.Children {
		if child.Value == value {
			return true
		}
	}

	return false
}
