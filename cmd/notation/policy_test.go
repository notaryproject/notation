package main

import (
	"reflect"
	"testing"
)

const (
	testPolicyName = "test_policy_name"
	testScope      = "test_scope"
	testPath = "test_path"
	testLevel = "test_level"
	testOverride = "key=val"
	testTrustStore = "ca:test"
	testIdentity = "testIdentity"
)

func TestPolicyListCommand_BasicArgs(t *testing.T) {
	cmd := policyListCommand()
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("Parse flag failed: %v", err)
	}
}

func TestPolicyShowCommand_ValidArgs(t *testing.T) {
	cmd := policyShowCommand()
	if err := cmd.ParseFlags([]string{testPolicyName}); err != nil {
		t.Fatalf("Parse flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
}

func TestPolicyShowCommand_NoArg(t *testing.T) {
	cmd := policyShowCommand()
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("Parse flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Error should be returned but got nil")
	}
}

func TestPolicyResolveCommand_ValidArgs(t *testing.T) {
	cmd := policyResolveCommand()
	if err := cmd.ParseFlags([]string{testScope}); err != nil {
		t.Fatalf("Parse flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
}

func TestPolicyResolveCommand_NoArgs(t *testing.T) {
	cmd := policyResolveCommand()
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("Parse flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Error should be returned but got nil")
	}
}

func TestPolicyDeleteCommand_ValidArgs(t *testing.T) {
	opts := &deleteOpts{}
	cmd := policyDeleteCommand(opts)
	expected := &deleteOpts{
		names:     []string{testPolicyName},
		confirmed: false,
	}
	if err := cmd.ParseFlags(expected.names); err != nil {
		t.Fatalf("Parse flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if !reflect.DeepEqual(opts, expected) {
		t.Fatalf("Expect opts: %+v, got: %+v", expected, opts)
	}
}

func TestPolicyDeleteCommand_NoArgs(t *testing.T) {
	opts := &deleteOpts{}
	cmd := policyDeleteCommand(opts)
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("Parse flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Error should be returned but got nil")
	}
}

func TestPolicyUpdateCommand_NoArgs(t *testing.T) {
	opts := &policyOpts{}
	cmd := policyUpdateCommand(opts)
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("Parse flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Error should be returned but got nil")
	}
}

func TestPolicyUpdateCommand_FilePath(t *testing.T) {
	opts := &policyOpts{}
	cmd := policyUpdateCommand(opts)
	expected := &policyOpts{
		configPath: testPath,
	}
	if err := cmd.ParseFlags([]string{testPath}); err != nil {
		t.Fatalf("Parse flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if !reflect.DeepEqual(opts, expected) {
		t.Fatalf("Expect opts: %+v, got: %+v", expected, opts)
	}
}

func TestPolicyUpdateCommand_Flags(t *testing.T) {
	opts := &policyOpts{}
	cmd := policyUpdateCommand(opts)
	expected := &policyOpts{
		configPath: testPolicyName,
		scopes: []string{testScope},
		level: testLevel,
		override: testOverride,
		stores: []string{testTrustStore},
		identities: []string{testIdentity},
	}
	if err := cmd.ParseFlags([]string{
		"--scope", testScope,
		"--level", testLevel,
		"--level-override", testOverride,
		"--trust-store", testTrustStore,
		"--identity", testIdentity,
		testPolicyName,
	}); err != nil {
		t.Fatalf("Parse flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if !reflect.DeepEqual(opts, expected) {
		t.Fatalf("Expect opts: %+v, got: %+v", expected, opts)
	}
}

func TestPolicyAddCommand_FilePath(t *testing.T) {
	opts := &policyOpts{}
	cmd := policyAddCommand(opts)
	expected := &policyOpts{
		configPath: testPath,
	}
	if err := cmd.ParseFlags([]string{testPath}); err != nil {
		t.Fatalf("Parse flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if !reflect.DeepEqual(opts, expected) {
		t.Fatalf("Expect opts: %+v, got: %+v", expected, opts)
	}
}

func TestPolicyAddCommand_Flags(t *testing.T) {
	opts := &policyOpts{}
	cmd := policyAddCommand(opts)
	expected := &policyOpts{
		configPath: testPolicyName,
		scopes: []string{testScope},
		level: testLevel,
		override: testOverride,
		stores: []string{testTrustStore},
		identities: []string{testIdentity},
		certPath: testPath,
	}
	if err := cmd.ParseFlags([]string{
		"--scope", testScope,
		"--level", testLevel,
		"--level-override", testOverride,
		"--trust-store", testTrustStore,
		"--identity", testIdentity,
		"--identity-cert", testPath,
		testPolicyName,
	}); err != nil {
		t.Fatalf("Parse flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse args failed: %v", err)
	}
	if !reflect.DeepEqual(opts, expected) {
		t.Fatalf("Expect opts: %+v, got: %+v", expected, opts)
	}
}