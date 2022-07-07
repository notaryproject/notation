package main

import (
	"fmt"
	"testing"
)

func TestPushCommand(t *testing.T) {
	cmd := pushCommand()
	cmd.ParseFlags([]string{"n0", "n1", "n2"})
	if narg := cmd.Flags().NArg(); narg != 3 {
		t.Fatalf("Expect reference number: %v, got: %v", 3, narg)
	}
	for i, arg := range cmd.Flags().Args() {
		if expected := fmt.Sprintf("n%v", i); arg != expected {
			t.Fatalf("Expect reference name: %v, got: %v", expected, arg)
		}
	}
}
