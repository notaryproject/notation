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

// Node represents a Node in a tree
type Node struct {
	Value    string
	Children []*Node
}

// New creates a new Node with the given value
func New(value string) *Node {
	return &Node{Value: value}
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

// Print prints the tree represented by the root node
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
