package tree

import (
	"bytes"
	"testing"
)

func TestStreamingPrinter(t *testing.T) {
	t.Run("one node", func(t *testing.T) {
		expected := "└── a\n"
		buff := &bytes.Buffer{}
		p := newStreamingPrinter("", buff)
		p.PrintNode(newNode("a"))
		p.Complete()

		if buff.String() != expected {
			t.Fatalf("expected %s, got %s", expected, buff.String())
		}
	})

	t.Run("two nodes", func(t *testing.T) {
		expected := `├── a
└── b
`
		buff := &bytes.Buffer{}
		p := newStreamingPrinter("", buff)
		p.PrintNode(newNode("a"))
		p.PrintNode(newNode("b"))
		p.Complete()

		if buff.String() != expected {
			t.Fatalf("expected %s, got %s", expected, buff.String())
		}
	})

	t.Run("two node with complex structure", func(t *testing.T) {
		expected := `├── a
│   ├── b
│   │   └── c
│   └── d
└── e
    ├── f
    │   └── g
    └── h
`
		buff := &bytes.Buffer{}
		p := newStreamingPrinter("", buff)
		// create the tree
		a := newNode("a")
		b := a.Add("b")
		b.Add("c")
		a.Add("d")
		p.PrintNode(a)

		e := newNode("e")
		f := e.Add("f")
		f.Add("g")
		e.Add("h")
		p.PrintNode(e)

		p.Complete()

		if buff.String() != expected {
			t.Fatalf("expected %s, got %s", expected, buff.String())
		}
	})

	t.Run("two node with prefix", func(t *testing.T) {
		expected := `    │   ├── a
    │   │   ├── b
    │   │   │   └── c
    │   │   └── d
    │   └── e
    │       ├── f
    │       │   └── g
    │       └── h
`
		buff := &bytes.Buffer{}
		p := newStreamingPrinter("    │   ", buff)
		// create the tree
		a := newNode("a")
		b := a.Add("b")
		b.Add("c")
		a.Add("d")
		p.PrintNode(a)

		e := newNode("e")
		f := e.Add("f")
		f.Add("g")
		e.Add("h")
		p.PrintNode(e)

		p.Complete()

		if buff.String() != expected {
			t.Fatalf("expected %s, got %s", expected, buff.String())
		}
	})
}
