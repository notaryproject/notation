package main

import (
	"fmt"
	"testing"
	"time"
)

func TestCertAddCommand(t *testing.T) {
	cmd := certAddCommand()
	err := cmd.ParseFlags([]string{
		"fff",
		"-n", "nnn"})
	if err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if arg := cmd.Flags().Arg(0); arg != "fff" {
		t.Fatalf("Expect cert: %v, got: %v", "fff", arg)
	}
	if val, _ := cmd.Flags().GetString("name"); val != "nnn" {
		t.Fatalf("Expect name: %v, got: %v", "nnn", val)
	}
}

func TestCertRemoveCommand(t *testing.T) {
	cmd := certRemoveCommand()
	err := cmd.ParseFlags([]string{
		"c0", "c1", "c2"})
	if err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if narg := cmd.Flags().NArg(); narg != 3 {
		t.Fatalf("Expect cert remove number: %v, got: %v", 3, narg)
	}
	for i, arg := range cmd.Flags().Args() {
		if expected := fmt.Sprintf("c%v", i); arg != expected {
			t.Fatalf("Expect cert: %v, got: %v", expected, arg)
		}
	}
}

func TestCertGenerateCommand(t *testing.T) {
	cmd := certGenerateTestCommand()
	err := cmd.ParseFlags([]string{
		"h0", "h1",
		"-n", "nnn",
		"--bits", "1024",
		"h2",
		"--default"})
	if err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if narg := cmd.Flags().NArg(); narg != 3 {
		t.Fatalf("Expect cert remove number: %v, got: %v", 3, narg)
	}
	for i, arg := range cmd.Flags().Args() {
		if expected := fmt.Sprintf("h%v", i); arg != expected {
			t.Fatalf("Expect host: %v, got: %v", expected, arg)
		}
	}
	if val, _ := cmd.Flags().GetBool("trust"); val {
		t.Fatalf("Expect trust: %v, got: %v", false, val)
	}
	if val, _ := cmd.Flags().GetInt("bits"); val != 1024 {
		t.Fatalf("Expect bits: %v, got: %v", 1024, val)
	}
	if val, _ := cmd.Flags().GetString("name"); val != "nnn" {
		t.Fatalf("Expect name: %v, got: %v", "nnn", val)
	}
	defaultExpiry := 365 * 24 * time.Hour
	if val, _ := cmd.Flags().GetDuration("expiry"); val != defaultExpiry {
		t.Fatalf("Expect expiry: %v, got: %v", defaultExpiry, val)
	}
	if val, _ := cmd.Flags().GetBool("default"); !val {
		t.Fatalf("Expect default: %v, got: %v", true, val)
	}
}
