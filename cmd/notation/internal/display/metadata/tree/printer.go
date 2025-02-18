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

import "io"

// streamingPrinter prints the tree nodes in a streaming fashion.
type streamingPrinter struct {
	w        io.Writer
	prefix   string
	prevNode *node
}

// newStreamingPrinter creates a new streaming printer.
//
// prefix is the prefix string that will be inherited by the nodes that are
// printed.
func newStreamingPrinter(prefix string, w io.Writer) *streamingPrinter {
	return &streamingPrinter{
		w:      w,
		prefix: prefix,
	}
}

// PrintNode adds a new node to be ready to print.
func (p *streamingPrinter) PrintNode(node *node) {
	if p.prevNode == nil {
		p.prevNode = node
		return
	}
	print(p.w, p.prefix, treeItemPrefix, p.prefix+subTreePrefix, p.prevNode)
	p.prevNode = node
}

// Complete prints the last node and completes the printing.
func (p *streamingPrinter) Complete() {
	if p.prevNode != nil {
		// print the last node
		print(p.w, p.prefix, treeItemPrefixLast, p.prefix+subTreePrefixLast, p.prevNode)
		p.prevNode = nil
	}
}
