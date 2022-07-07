package main

import (
	"testing"
)

func TestPullCommand(t *testing.T) {
	cmd := pullCommand()
	cmd.ParseFlags([]string{"n0"})
	if val := cmd.Flags().Arg(0); val != "n0" {
		t.Fatalf("Expect reference name: %v, got: %v", "n0", val)
	}
}
