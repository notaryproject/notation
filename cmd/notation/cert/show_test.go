package cert

import (
	"reflect"
	"testing"
)

func TestCertShowCommand(t *testing.T) {
	opts := &certShowOpts{}
	cmd := certShowCommand(opts)
	expected := &certShowOpts{
		storeType:  "ca",
		namedStore: "test",
		cert:       "test.crt",
	}
	if err := cmd.ParseFlags([]string{
		"test.crt",
		"-t", "ca",
		"-s", "test"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect cert show opts: %v, got: %v", expected, opts)
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
