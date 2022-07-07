package main

import (
	"testing"
	"time"

	"github.com/notaryproject/notation/internal/cmd"
)

func TestSignCommand(t *testing.T) {
	command := signCommand()
	err := command.ParseFlags([]string{
		"ref",
		"--key", "kkk",
		"--key-file", "fff",
		"--cert-file", "ccc",
		"-r", "ref",
		"--origin",
		"--timestamp", "0000",
		"--expiry", "365s"})
	if err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if ref := command.Flags().Arg(0); ref != "ref" {
		t.Fatalf("Expect reference: %v, got: %v", "ref", ref)
	}
	if name, _ := command.Flags().GetString(cmd.FlagKey.Name); name != "kkk" {
		t.Fatalf("Expect %v: %v, got: %v", cmd.FlagKey.Name, "kkk", name)
	}
	if keyFile, _ := command.Flags().GetString(cmd.FlagKeyFile.Name); keyFile != "fff" {
		t.Fatalf("Expect %v: %v, got: %v", cmd.FlagKeyFile.Name, "fff", keyFile)
	}
	if certFile, _ := command.Flags().GetString(cmd.FlagCertFile.Name); certFile != "ccc" {
		t.Fatalf("Expect %v: %v, got: %v", cmd.FlagCertFile.Name, "ccc", certFile)
	}
	if origin, _ := command.Flags().GetBool("origin"); !origin {
		t.Fatalf("Expect %v: %v, got: %v", "origin", true, origin)
	}
	if tm, _ := command.Flags().GetString(cmd.FlagTimestamp.Name); tm != "0000" {
		t.Fatalf("Expect %v: %v, got: %v", cmd.FlagTimestamp.Name, "0000", tm)
	}
	if expiry, _ := command.Flags().GetDuration(cmd.FlagExpiry.Name); expiry != time.Second*365 {
		t.Fatalf("Expect %v: %v, got: %v", cmd.FlagExpiry.Name, time.Second*365, expiry)
	}
}
