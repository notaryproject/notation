package cert

import (
	"reflect"
	"testing"
)

func TestCertAddCommand(t *testing.T) {
	opts := &certAddOpts{}
	cmd := certAddCommand(opts)
	expected := &certAddOpts{
		storeType:  "ca",
		namedStore: "test",
		path:       []string{"path"},
	}
	if err := cmd.ParseFlags([]string{
		"path",
		"-t", "ca",
		"-s", "test"}); err != nil {
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
