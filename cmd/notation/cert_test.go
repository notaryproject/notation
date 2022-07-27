package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestCertAddCommand(t *testing.T) {
	opts := &certAddOpts{}
	cmd := certAddCommand(opts)
	expected := &certAddOpts{
		path: "path",
		name: "cert",
	}
	if err := cmd.ParseFlags([]string{
		expected.path,
		"-n", expected.name}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if *expected != *opts {
		t.Fatalf("Expect cert add opts: %v, got: %v", expected, opts)
	}
}

func TestCertAddCommand_MissinArgs(t *testing.T) {
	cmd := certAddCommand(nil)
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}

func TestCertRemoveCommand(t *testing.T) {
	opts := &certRemoveOpts{}
	cmd := certRemoveCommand(opts)
	expected := &certRemoveOpts{
		names: []string{"cert0", "cert1", "cert2"},
	}
	if err := cmd.ParseFlags(expected.names); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect cert remove opts: %v, got: %v", expected, opts)
	}
}

func TestCertRemoveCommand_MissinArgs(t *testing.T) {
	cmd := certRemoveCommand(nil)
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}

func TestCertGenerateCommand(t *testing.T) {
	opts := &certGenerateTestOpts{}
	cmd := certGenerateTestCommand(opts)
	expected := &certGenerateTestOpts{
		hosts:     []string{"host0", "host1", "host2"},
		name:      "name",
		bits:      2048,
		isDefault: true,
		expiry:    365 * 24 * time.Hour,
	}
	if err := cmd.ParseFlags([]string{
		"host0", "host1",
		"-n", expected.name,
		"--bits", fmt.Sprint(expected.bits),
		"host2",
		"--default"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect cert generate test opts: %v, got: %v", expected, opts)
	}
}

func TestCertGenerateTestCommand_MissinArgs(t *testing.T) {
	cmd := certGenerateTestCommand(nil)
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}
