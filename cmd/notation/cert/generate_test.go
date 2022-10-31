package cert

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCertGenerateCommand(t *testing.T) {
	opts := &certGenerateTestOpts{}
	cmd := certGenerateTestCommand(opts)
	expected := &certGenerateTestOpts{
		name:      "name",
		bits:      2048,
		isDefault: true,
	}
	if err := cmd.ParseFlags([]string{
		"name",
		"--bits", fmt.Sprint(expected.bits),
		"--default"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect cert generate-test opts: %v, got: %v", expected, opts)
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
