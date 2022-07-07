package main

import (
	"fmt"
	"os"
	"testing"
)

func TestListCommand(t *testing.T) {
	cmd := listCommand()
	err := cmd.ParseFlags([]string{
		"--password", "ppp",
		"n0",
		"-u", "uuu",
		"n1", "n2",
		"--plain-http"})
	if err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if narg := cmd.Flags().NArg(); narg != 3 {
		t.Fatalf("Expect reference number: %v, got: %v", 3, narg)
	}
	for i, arg := range cmd.Flags().Args() {
		if expected := fmt.Sprintf("n%v", i); arg != expected {
			t.Fatalf("Expect reference name: %v, got: %v", expected, arg)
		}
	}
	if name, _ := cmd.Flags().GetString(flagUsername.Name); name != "uuu" {
		t.Fatalf("Expect %v: %v, got: %v", flagUsername.Name, "uuu", name)
	}
	if plain, _ := cmd.Flags().GetBool(flagPlainHTTP.Name); !plain {
		t.Fatalf("Expect %v: %v, got: %v", flagPlainHTTP.Name, true, plain)
	}
	if password, _ := cmd.Flags().GetString(flagPassword.Name); password != "ppp" {
		t.Fatalf("Expect %v: %v, got: %v", flagPassword.Name, "ppp", password)
	}
}

func TestListCommandFromEnv(t *testing.T) {
	os.Setenv(defaultUsernameEnv, "uuu")
	os.Setenv(defaultPasswordEnv, "ppp")
	defer os.Unsetenv(defaultUsernameEnv)
	defer os.Unsetenv(defaultPasswordEnv)
	cmd := listCommand()
	err := cmd.ParseFlags([]string{
		"n0", "n1", "n2"})
	if err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	for i, arg := range cmd.Flags().Args() {
		if expected := fmt.Sprintf("n%v", i); arg != expected {
			t.Fatalf("Expect reference name: %v, got: %v", expected, arg)
		}
	}
	if name, _ := cmd.Flags().GetString(flagUsername.Name); name != "uuu" {
		t.Fatalf("Expect %v: %v, got: %v", flagUsername.Name, "uuu", name)
	}
	if password, _ := cmd.Flags().GetString(flagPassword.Name); password != "ppp" {
		t.Fatalf("Expect %v: %v, got: %v", flagPassword.Name, "ppp", password)
	}
}
