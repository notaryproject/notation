package main

import "testing"

func TestGenerateCommand(t *testing.T) {
	cmd := generateCommand()
	subCmds := cmd.Commands()
	if len(subCmds) != 1 {
		t.Fatalf("Expect generate command have 1 subcommand, got: %v", len(subCmds))
	}
}
