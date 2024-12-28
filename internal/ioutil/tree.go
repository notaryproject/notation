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
		printNodeDetails(mainChild, subTreePrefixLast)
	}

	return nil
}

// PrintKeyMap prints the given key map in a tree format
func printNodeDetails(node *Node, prefix string) {
	// Print pairs
	if node.Pairs != nil {
		for key, value := range node.Pairs {
			fmt.Printf("%s%s%s: %s\n", prefix, treeItemPrefix, key, value)
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
	if node.Pairs != nil {
		for key, value := range node.Pairs {
			if value == "(empty)" {
				fmt.Printf("%s    %s\n", prefix, value)
			} else {
				fmt.Printf("%s    %s: %s\n", prefix, key, value)
			}
		}
	}

	for _, child := range node.Children {
		fmt.Printf("%s    %s\n", prefix, child.Value)
		childPrefix := prefix + subTreePrefixLast
		printNodeDetails(child, childPrefix)
	}
}
