package main

import (
	"fmt"
	"testing"
)

func TestCacheListCommand(t *testing.T) {
	cmd := cacheListCommand()
	err := cmd.ParseFlags([]string{
		"-l", "lll",
		"mmm",
		"-u", "uuu"})
	if err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if arg := cmd.Flags().Arg(0); arg != "mmm" {
		t.Fatalf("Expect manifest: %v, got: %v", "mmm", arg)
	}
	if name, _ := cmd.Flags().GetString(flagUsername.Name); name != "uuu" {
		t.Fatalf("Expect %v: %v, got: %v", flagUsername.Name, "uuu", name)
	}
	if local, _ := cmd.Flags().GetString(flagLocal.Name); local != "lll" {
		t.Fatalf("Expect %v: %v, got: %v", flagLocal.Name, "lll", local)
	}
}

func TestCachePruneCommand(t *testing.T) {
	cmd := cachePruneCommand()
	err := cmd.ParseFlags([]string{
		"-a",
		"--purge",
		"ref0", "ref1", "ref2",
		"-f"})
	if err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if narg := cmd.Flags().NArg(); narg != 3 {
		t.Fatalf("Expect reference number: %v, got: %v", 3, narg)
	}
	for i, arg := range cmd.Flags().Args() {
		if expected := fmt.Sprintf("ref%v", i); arg != expected {
			t.Fatalf("Expect reference: %v, got: %v", expected, arg)
		}
	}
	if val, _ := cmd.Flags().GetBool("all"); !val {
		t.Fatalf("Expect all: %v, got: %v", true, val)
	}
	if val, _ := cmd.Flags().GetBool("purge"); !val {
		t.Fatalf("Expect purge: %v, got: %v", true, val)
	}
	if val, _ := cmd.Flags().GetBool("force"); !val {
		t.Fatalf("Expect force: %v, got: %v", true, val)
	}
}

func TestCacheRemoveCommand(t *testing.T) {
	cmd := cacheRemoveCommand()
	err := cmd.ParseFlags([]string{
		"--plain-http",
		"ref",
		"digest0", "digest1",
		"--password", "ppp"})
	if err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if narg := cmd.Flags().NArg(); narg != 3 {
		t.Fatalf("Expect reference number: %v, got: %v", 3, narg)
	}
	if ref := cmd.Flags().Arg(0); ref != "ref" {
		t.Fatalf("Expect reference: %v, got: %v", "ref", ref)
	}
	for i, arg := range cmd.Flags().Args()[1:] {
		if expected := fmt.Sprintf("digest%v", i); arg != expected {
			t.Fatalf("Expect digest: %v, got: %v", expected, arg)
		}
	}
	if plain, _ := cmd.Flags().GetBool(flagPlainHTTP.Name); !plain {
		t.Fatalf("Expect %v: %v, got: %v", flagPlainHTTP.Name, true, plain)
	}
	if password, _ := cmd.Flags().GetString(flagPassword.Name); password != "ppp" {
		t.Fatalf("Expect %v: %v, got: %v", flagPassword.Name, "ppp", password)
	}
}
