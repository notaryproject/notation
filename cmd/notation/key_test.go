package main

import (
	"fmt"
	"testing"
)

func TestKeyAddCommand(t *testing.T) {
	cmd := keyAddCommand()
	err := cmd.ParseFlags([]string{
		"-n", "nnn",
		"--plugin", "ppp",
		"--id", "iii",
		"kkk", "ccc"})
	if err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if arg := cmd.Flags().Arg(0); arg != "kkk" {
		t.Fatalf("Expect key_path: %v, got: %v", "kkk", arg)
	}
	if arg := cmd.Flags().Arg(1); arg != "ccc" {
		t.Fatalf("Expect cert_path: %v, got: %v", "ccc", arg)
	}
	if val, _ := cmd.Flags().GetString("name"); val != "nnn" {
		t.Fatalf("Expect name: %v, got: %v", "nnn", val)
	}
	if val, _ := cmd.Flags().GetString("plugin"); val != "ppp" {
		t.Fatalf("Expect plugin: %v, got: %v", "ppp", val)
	}
	if val, _ := cmd.Flags().GetString("id"); val != "iii" {
		t.Fatalf("Expect plugin id: %v, got: %v", "iii", val)
	}
}

func TestKeyUpdateCommand(t *testing.T) {
	cmd := keyUpdateCommand()
	err := cmd.ParseFlags([]string{
		"n0", "n1", "n2",
		"--default"})
	if err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if narg := cmd.Flags().NArg(); narg != 3 {
		t.Fatalf("Expect key update number: %v, got: %v", 3, narg)
	}
	for i, arg := range cmd.Flags().Args() {
		if expected := fmt.Sprintf("n%v", i); arg != expected {
			t.Fatalf("Expect key name: %v, got: %v", expected, arg)
		}
	}
	if val, _ := cmd.Flags().GetBool("default"); !val {
		t.Fatalf("Expect default: %v, got: %v", true, val)
	}
}

func TestKeyRemoveCommand(t *testing.T) {
	cmd := keyRemoveCommand()
	err := cmd.ParseFlags([]string{
		"n0", "n1", "n2"})
	if err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if narg := cmd.Flags().NArg(); narg != 3 {
		t.Fatalf("Expect key remove number: %v, got: %v", 3, narg)
	}
	for i, arg := range cmd.Flags().Args() {
		if expected := fmt.Sprintf("n%v", i); arg != expected {
			t.Fatalf("Expect key name: %v, got: %v", expected, arg)
		}
	}
}
