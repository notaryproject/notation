package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCertAddCommand_BasicArgs(t *testing.T) {
	opts := &certAddOpts{}
	cmd := certAddCommand(opts)
	expected := &certAddOpts{
		storeType:  "ca",
		namedStore: "test1",
		path:       []string{"cert.pem"},
	}
	if err := cmd.ParseFlags([]string{
		"cert.pem",
		"-t", "ca", "-s", "test1"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect cert add opts: %v, got: %v", expected, opts)
	}
}

func TestCertAddCommand_MissingArgs(t *testing.T) {
	cmd := certAddCommand(nil)
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}

func TestCertListCommand(t *testing.T) {
	opts := &certListOpts{}
	cmd := certListCommand(opts)
	expected := &certListOpts{
		storeType:  "ca",
		namedStore: "test1",
	}
	if err := cmd.ParseFlags([]string{
		"-t", "ca", "-s", "test1"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect cert add opts: %v, got: %v", expected, opts)
	}
}

func TestCertShowCommand_BasicArgs(t *testing.T) {
	opts := &certShowOpts{}
	cmd := certShowCommand(opts)
	expected := &certShowOpts{
		storeType:  "ca",
		namedStore: "test1",
		cert:       "cert.pem",
	}
	if err := cmd.ParseFlags([]string{
		"cert.pem",
		"-t", "ca", "-s", "test1"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect cert add opts: %v, got: %v", expected, opts)
	}
}

func TestCertShowCommand_MissingArgs(t *testing.T) {
	cmd := certShowCommand(nil)
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}

func TestCertRemoveCommand_BasicArgs(t *testing.T) {
	opts := &certRemoveOpts{}
	cmd := certRemoveCommand(opts)
	expected := &certRemoveOpts{
		storeType:  "ca",
		namedStore: "test1",
		all:        true,
	}
	if err := cmd.ParseFlags([]string{
		"-t", "ca", "-s", "test1", "-a"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect cert remove opts: %v, got: %v", expected, opts)
	}
}

func TestCertRemoveCommand_MissingArgs(t *testing.T) {
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
		host:      "host0",
		name:      "name",
		bits:      2048,
		isDefault: true,
	}
	if err := cmd.ParseFlags([]string{
		"host0",
		"-n", "name",
		"--bits", fmt.Sprint(2048),
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

func TestCertGenerateTestCommand_MissingArgs(t *testing.T) {
	cmd := certGenerateTestCommand(nil)
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}
