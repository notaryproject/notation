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
	"io"
)

const (
	treeItemPrefix     = "├── "
	treeItemPrefixLast = "└── "
	subTreePrefix      = "│   "
	subTreePrefixLast  = "    "
)

// represents a node in a tree
type node struct {
	Value    string
	Children []*node
}

// creates a newNode node with the given value
func newNode(value string) *node {
	return &node{Value: value}
}

// adds a new child node with the given value
func (parent *node) Add(value string) *node {
	node := newNode(value)
	parent.Children = append(parent.Children, node)
	return node
}

// adds a new child node with the formatted pair as the value
func (parent *node) AddPair(key string, value string) *node {
	return parent.Add(key + ": " + value)
}

// prints the tree represented by the root node
func (root *node) Print(w io.Writer) error {
	return print(w, "", "", "", root)
}

func print(w io.Writer, prefix string, itemMarker string, nextPrefix string, n *node) error {
	if _, err := fmt.Fprintln(w, prefix+itemMarker+n.Value); err != nil {
		return err
	}

	nextItemPrefix := treeItemPrefix
	nextSubTreePrefix := subTreePrefix

	if len(n.Children) > 0 {
		for i, child := range n.Children {
			if i == len(n.Children)-1 {
				nextItemPrefix = treeItemPrefixLast
				nextSubTreePrefix = subTreePrefixLast
			}
			if err := print(w, nextPrefix, nextItemPrefix, nextPrefix+nextSubTreePrefix, child); err != nil {
				return err
			}
		}
	}

	return nil
}
