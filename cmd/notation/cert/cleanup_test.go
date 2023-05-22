package cert

import (
	"reflect"
	"testing"
)

func TestCertCleanupTestCommand(t *testing.T) {
	opts := &certCleanupTestOpts{}
	cmd := certCleanupTestCommand(opts)
	expected := &certCleanupTestOpts{
		keyName: "name",
	}
	if err := cmd.ParseFlags([]string{"name"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect cert generate-test opts: %v, got: %v", expected, opts)
	}
}

func TestCertCleanupTestCommand_MissingArgs(t *testing.T) {
	cmd := certCleanupTestCommand(nil)
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}
