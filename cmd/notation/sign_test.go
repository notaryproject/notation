package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/notaryproject/notation/internal/cmd"
)

func TestSignCommand(t *testing.T) {
	command := signCommand()
	err := command.ParseFlags([]string{
		"ref",
		"-u", "uuu",
		"--password", "ppp",
		"--key", "kkk",
		"--key-file", "fff"})
	if err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if ref := command.Flags().Arg(0); ref != "ref" {
		t.Fatalf("Expect reference: %v, got: %v", "ref", ref)
	}
	if name, _ := command.Flags().GetString(flagUsername.Name); name != "uuu" {
		t.Fatalf("Expect %v: %v, got: %v", flagUsername.Name, "uuu", name)
	}
	if key, _ := command.Flags().GetString(cmd.FlagKey.Name); key != "kkk" {
		t.Fatalf("Expect %v: %v, got: %v", cmd.FlagKey.Name, "kkk", key)
	}
	if keyFile, _ := command.Flags().GetString(cmd.FlagKeyFile.Name); keyFile != "fff" {
		t.Fatalf("Expect %v: %v, got: %v", cmd.FlagKeyFile.Name, "fff", keyFile)
	}
	if plain, _ := command.Flags().GetBool(flagPlainHTTP.Name); plain {
		t.Fatalf("Expect %v: %v, got: %v", flagPlainHTTP.Name, false, plain)
	}
	if password, _ := command.Flags().GetString(flagPassword.Name); password != "ppp" {
		t.Fatalf("Expect %v: %v, got: %v", flagPassword.Name, "ppp", password)
	}
	if push, _ := command.Flags().GetBool("push"); !push {
		t.Fatalf("Expect push: %v, got: %v", true, false)
	}
	if mediaType, _ := command.Flags().GetString(flagMediaType.Name); mediaType != defaultMediaType {
		t.Fatalf("Expect %v: %v, got: %v", flagMediaType.Name, defaultMediaType, mediaType)
	}
}

func TestSignWithMoreFlagCommand(t *testing.T) {
	command := signCommand()
	err := command.ParseFlags([]string{
		"ref",
		"--key", "kkk",
		"--key-file", "fff",
		"--plain-http",
		"--push=false",
		"--media-type", "mmm",
		"-l", "lll",
		"--output", "ooo",
		"--expiry", "365s"})

	if err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if ref := command.Flags().Arg(0); ref != "ref" {
		t.Fatalf("Expect reference: %v, got: %v", "ref", ref)
	}
	if key, _ := command.Flags().GetString(cmd.FlagKey.Name); key != "kkk" {
		t.Fatalf("Expect %v: %v, got: %v", cmd.FlagKey.Name, "kkk", key)
	}
	if keyFile, _ := command.Flags().GetString(cmd.FlagKeyFile.Name); keyFile != "fff" {
		t.Fatalf("Expect %v: %v, got: %v", cmd.FlagKeyFile.Name, "fff", keyFile)
	}
	if plain, _ := command.Flags().GetBool(flagPlainHTTP.Name); !plain {
		t.Fatalf("Expect %v: %v, got: %v", flagPlainHTTP.Name, true, plain)
	}
	if push, _ := command.Flags().GetBool("push"); push {
		t.Fatalf("Expect push: %v, got: %v", false, true)
	}
	if mediaType, _ := command.Flags().GetString(flagMediaType.Name); mediaType != "mmm" {
		t.Fatalf("Expect %v: %v, got: %v", flagMediaType.Name, "mmm", mediaType)
	}
	if local, _ := command.Flags().GetString(flagLocal.Name); local != "lll" {
		t.Fatalf("Expect %v: %v, got: %v", flagLocal.Name, "lll", local)
	}
	if output, _ := command.Flags().GetString(flagOutput.Name); output != "ooo" {
		t.Fatalf("Expect %v: %v, got: %v", flagOutput.Name, "ooo", output)
	}
	if expiry, _ := command.Flags().GetDuration(cmd.FlagExpiry.Name); expiry != time.Second*365 {
		t.Fatalf("Expect %v: %v, got: %v", cmd.FlagExpiry.Name, time.Second*365, expiry)
	}
}

func TestSignWithConfig(t *testing.T) {
	command := signCommand()
	command.ParseFlags([]string{"-c", "key0=val0,key1=val1,key2=val2"})
	config, err := cmd.ParseFlagPluginConfig(command)
	if err != nil {
		t.Fatalf("Parse plugin Config flag failed: %v", err)
	}
	if len(config) != 3 {
		t.Fatalf("Expect plugin config number: %v, got: %v ", 3, len(config))
	}
	for i := 0; i < 3; i++ {
		key, val := fmt.Sprintf("key%v", i), fmt.Sprintf("val%v", i)
		configVal, ok := config[key]
		if !ok {
			t.Fatalf("Key: %v not in config", key)
		}
		if val != configVal {
			t.Fatalf("Value for key: %v error, got: %v, expect: %v", key, configVal, val)
		}
	}
}
