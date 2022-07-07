package main

import (
	"os"
	"testing"
)

func TestPullCommand(t *testing.T) {
	os.Setenv(defaultUsernameEnv, "uuu")
	os.Setenv(defaultPasswordEnv, "ppp")
	defer os.Unsetenv(defaultUsernameEnv)
	defer os.Unsetenv(defaultPasswordEnv)

	cmd := pullCommand()
	err := cmd.ParseFlags([]string{
		"ref",
		"-u", "uuu2",
		"--strict"})
	if err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if ref := cmd.Flags().Arg(0); ref != "ref" {
		t.Fatalf("Expect reference: %v, got: %v", "ref", ref)
	}
	if name, _ := cmd.Flags().GetString(flagUsername.Name); name != "uuu2" {
		t.Fatalf("Expect %v: %v, got: %v", flagUsername.Name, "uuu2", name)
	}
	if plain, _ := cmd.Flags().GetBool(flagPlainHTTP.Name); plain {
		t.Fatalf("Expect %v: %v, got: %v", flagPlainHTTP.Name, false, plain)
	}
	if strict, _ := cmd.Flags().GetBool("strict"); !strict {
		t.Fatalf("Expect strict: %v, got: %v", true, strict)
	}
	if password, _ := cmd.Flags().GetString(flagPassword.Name); password != "ppp" {
		t.Fatalf("Expect %v: %v, got: %v", flagPassword.Name, "ppp", password)
	}
}
