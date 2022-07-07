package main

import (
	"fmt"
	"os"
	"testing"
)

func TestPushCommand(t *testing.T) {
	os.Setenv(defaultUsernameEnv, "uuu")
	os.Setenv(defaultPasswordEnv, "ppp")
	defer os.Unsetenv(defaultUsernameEnv)
	defer os.Unsetenv(defaultPasswordEnv)

	cmd := pushCommand()
	err := cmd.ParseFlags([]string{
		"ref",
		"-u", "uuu2",
		"--signature", "s0",
		"-s", "s1"})
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
	signatures, _ := cmd.Flags().GetStringSlice("signature")
	if len(signatures) != 2 {
		t.Fatalf("Expect signarure number: %v, got: %v", 2, len(signatures))
	}
	for i, signature := range signatures {
		if expected := fmt.Sprintf("s%v", i); expected != signature {
			t.Fatalf("Expect signature: %v, got: %v", expected, signature)
		}
	}
	if password, _ := cmd.Flags().GetString(flagPassword.Name); password != "ppp" {
		t.Fatalf("Expect %v: %v, got: %v", flagPassword.Name, "ppp", password)
	}
}
