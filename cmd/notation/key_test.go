package main

import (
	"reflect"
	"testing"
)

func TestKeyAddCommand(t *testing.T) {
	opts := &keyAddOpts{}
	cmd := keyAddCommand(opts)
	expected := &keyAddOpts{
		name:         "name",
		plugin:       "pluginname",
		id:           "pluginid",
		keyPath:      "keypath",
		certPath:     "certpath",
		pluginConfig: "pluginconfig",
	}
	if err := cmd.ParseFlags([]string{
		"-n", expected.name,
		"--plugin", expected.plugin,
		"--id", expected.id,
		"-c", expected.pluginConfig,
		expected.keyPath, expected.certPath}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if *expected != *opts {
		t.Fatalf("Expect key add opts: %v, got: %v", expected, opts)
	}
}

func TestKeyUpdateCommand(t *testing.T) {
	opts := &keyUpdateOpts{}
	cmd := keyUpdateCommand(opts)
	expected := &keyUpdateOpts{
		name:      "name",
		isDefault: true,
	}
	if err := cmd.ParseFlags([]string{
		expected.name,
		"--default"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if *expected != *opts {
		t.Fatalf("Expect key update opts: %v, got: %v", expected, opts)
	}
}

func TestKeyUpdateCommand_MissinArgs(t *testing.T) {
	cmd := keyUpdateCommand(nil)
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}

func TestKeyRemoveCommand(t *testing.T) {
	opts := &keyRemoveOpts{}
	cmd := keyRemoveCommand(opts)
	expected := &keyRemoveOpts{
		names: []string{"key0", "key1", "key2"},
	}
	if err := cmd.ParseFlags(expected.names); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect key remove opts: %v, got: %v", expected, opts)
	}
}

func TestKeyRemoveCommand_MissinArgs(t *testing.T) {
	cmd := keyRemoveCommand(nil)
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}
