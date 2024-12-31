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

package ioutil

import (
	"fmt"
	"strings"
)

const (
	treeItemPrefix     = "├── "
	treeItemPrefixLast = "└── "
	subTreePrefix      = "│   "
	subTreePrefixLast  = "    "
)

// Node represents a Node in a tree
type Node struct {
	Value    string
	Pairs    map[string]string
	Children []*Node
}

// New creates a new Node with the given value
func New(value string) *Node {
	return &Node{
		Value:    value,
		Children: make([]*Node, 0),
		Pairs:    make(map[string]string),
	}
}

// Add adds a new child node with the given value
func (parent *Node) Add(value string) *Node {
	node := New(value)
	parent.Children = append(parent.Children, node)
	return node
}

// AddPair adds a new child node with the formatted pair as the value
func (parent *Node) AddPair(key string, value string) *Node {
	return parent.Add(key + ": " + value)
}

// PrintObjectAsTree prints the given object in a tree format with detailed node information
func PrintObjectAsTree(root *Node) error {
	if root == nil {
		return nil
	}

	fmt.Println(root.Value)

	for i, mainChild := range root.Children {
		isLast := i == len(root.Children)-1
		prefix := treeItemPrefixLast
		if !isLast {
			prefix = treeItemPrefix
		}
		fmt.Printf("%s%s\n", prefix, mainChild.Value)

		// If there is no associated signature (this is associated to --oci-layout with no signature cmd)
		if len(mainChild.Children) == 0 {
			fmt.Printf("%s has no associated signature\n", mainChild.Value)
			continue
		}

		// For root level, use subTreePrefixLast for the last child
		initialPrefix := ""
		if isLast {
			initialPrefix = subTreePrefixLast
		} else {
			initialPrefix = subTreePrefix
		}
		printNodeDetails(mainChild, initialPrefix)
	}

	return nil
}

// PrintKeyMap prints the given key map in a tree format
func printNodeDetails(node *Node, prefix string) {
	// Print pairs
	if node.Pairs != nil {
		lastIdx := len(node.Pairs) - 1
		i := 0
		for key, value := range node.Pairs {
			isLast := i == lastIdx
			connector := treeItemPrefix
			if isLast {
				connector = treeItemPrefixLast
			}
			fmt.Printf("%s%s%s: %s\n", prefix, connector, key, value)
			i++
		}
	}

	// Process children
	for i, child := range node.Children {
		isChildLast := i == len(node.Children)-1
		connector := treeItemPrefix
		if isChildLast {
			connector = treeItemPrefixLast
		}

		fmt.Printf("%s%s%s\n", prefix, connector, child.Value)

		nextPrefix := prefix
		if isChildLast {
			nextPrefix += subTreePrefixLast
		} else {
			nextPrefix += subTreePrefix
		}

		if strings.Contains(child.Value, "attributes") || strings.Contains(child.Value, "certificates") {
			handleSpecialNode(child, nextPrefix)
		} else {
			printNodeDetails(child, nextPrefix)
		}
	}
}

// handleSpecialNode handles special nodes that require special formatting when printing
func handleSpecialNode(node *Node, prefix string) {
	pairsList := make([]struct {
		key   string
		value string
	}, 0)

	// Collect pairs in a slice to handle last item properly
	if node.Pairs != nil {
		for key, value := range node.Pairs {
			if value == "(empty)" {
				pairsList = append(pairsList, struct {
					key   string
					value string
				}{key: "", value: "(empty)"})
			} else {
				pairsList = append(pairsList, struct {
					key   string
					value string
				}{key: key, value: value})
			}
		}
	}

	// Print pairs with proper tree structure
	for i, pair := range pairsList {
		isLast := i == len(pairsList)-1 && len(node.Children) == 0
		connector := treeItemPrefix
		if isLast {
			connector = treeItemPrefixLast
		}

		if pair.value == "(empty)" {
			fmt.Printf("%s%s%s\n", prefix, connector, pair.value)
		} else {
			fmt.Printf("%s%s%s: %s\n", prefix, connector, pair.key, pair.value)
		}
	}

	// Process children if any
	for i, child := range node.Children {
		isLast := i == len(node.Children)-1
		connector := treeItemPrefix
		if isLast {
			connector = treeItemPrefixLast
		}

		fmt.Printf("%s%s%s\n", prefix, connector, child.Value)

		nextPrefix := prefix
		if isLast {
			nextPrefix += subTreePrefixLast
		} else {
			nextPrefix += subTreePrefix
		}
		printNodeDetails(child, nextPrefix)
	}
}
