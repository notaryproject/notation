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
	"fmt"
)

const (
	treeItemPrefix     = "├── "
	treeItemPrefixLast = "└── "
	subTreePrefix      = "│   "
	subTreePrefixLast  = "    "
)

// represents a Node in a tree
type Node struct {
	Value    string
	Children []*Node
}

// creates a new Node with the given value
func New(value string) *Node {
	return &Node{Value: value}
}

// adds a new child node with the given value
func (parent *Node) Add(value string) *Node {
	node := New(value)
	parent.Children = append(parent.Children, node)
	return node
}

// adds a new child node with the formatted pair as the value
func (parent *Node) AddPair(key string, value string) *Node {
	return parent.Add(key + ": " + value)
}

// prints the tree represented by the root node
func (root *Node) Print() {
	print("", "", "", root)
}

func print(prefix string, itemMarker string, nextPrefix string, n *Node) {
	fmt.Println(prefix + itemMarker + n.Value)

	nextItemPrefix := treeItemPrefix
	nextSubTreePrefix := subTreePrefix

	if len(n.Children) > 0 {
		for i, child := range n.Children {
			if i == len(n.Children)-1 {
				nextItemPrefix = treeItemPrefixLast
				nextSubTreePrefix = subTreePrefixLast
			}
			print(nextPrefix, nextItemPrefix, nextPrefix+nextSubTreePrefix, child)
		}
	}
}
