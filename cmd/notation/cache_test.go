package main

import (
	"reflect"
	"testing"
)

func TestCacheListCommand(t *testing.T) {
	opts := &cacheListOpts{}
	cmd := cacheListCommand(opts)
	expected := &cacheListOpts{
		RemoteFlagOpts: RemoteFlagOpts{
			SecureFlagOpts: SecureFlagOpts{
				Username: "user",
			},
			CommonFlagOpts: CommonFlagOpts{
				Local:     true,
				MediaType: defaultMediaType,
			},
		},
		reference: "ref",
	}
	if err := cmd.ParseFlags([]string{
		"-l",
		expected.reference,
		"-u", expected.Username}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if *expected != *opts {
		t.Fatalf("Expect cache list opts: %v, got: %v", expected, opts)
	}
}

func TestCachePruneCommand(t *testing.T) {
	opts := &cachePruneOpts{}
	cmd := cachePruneCommand(opts)
	expected := &cachePruneOpts{
		RemoteFlagOpts: RemoteFlagOpts{
			CommonFlagOpts: CommonFlagOpts{
				MediaType: defaultMediaType,
			},
		},
		all:        true,
		force:      true,
		purge:      true,
		references: []string{"ref0", "ref1", "ref2"},
	}
	if err := cmd.ParseFlags([]string{
		"-a",
		"--purge",
		"ref0", "ref1", "ref2",
		"-f"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect cache prune opts: %v, got: %v", expected, opts)
	}
}

func TestCachePruneCommand_MissinArgs(t *testing.T) {
	cmd := cachePruneCommand(nil)
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}

func TestCacheRemoveCommand(t *testing.T) {
	opts := &cacheRemoveOpts{}
	cmd := cacheRemoveCommand(opts)
	expected := &cacheRemoveOpts{
		RemoteFlagOpts: RemoteFlagOpts{
			CommonFlagOpts: CommonFlagOpts{
				MediaType: defaultMediaType,
			},
			SecureFlagOpts: SecureFlagOpts{
				Password:  "password",
				PlainHTTP: true,
			},
		},
		reference:  "ref",
		sigDigests: []string{"digest0", "digest1"},
	}
	if err := cmd.ParseFlags([]string{
		"--plain-http",
		"ref",
		"digest0", "digest1",
		"--password", expected.Password}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect cache remove opts: %v, got: %v", expected, opts)
	}
}

func TestCacheRemoveCommand_MissinArgs(t *testing.T) {
	cmd := cacheRemoveCommand(nil)
	if err := cmd.ParseFlags([]string{"reference"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}
