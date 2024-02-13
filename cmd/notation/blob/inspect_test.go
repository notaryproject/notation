package blob

import (
	"github.com/notaryproject/notation/internal/cmd"
	"reflect"
	"testing"
)

func TestBlobInspectCommand_SecretsFromArgs(t *testing.T) {
	opts := &blobInspectOpts{}
	command := inspectCommand(opts)
	expected := &blobInspectOpts{
		signaturePath: "path",
		outputFormat:  cmd.OutputPlaintext}
	if err := command.ParseFlags([]string{
		expected.signaturePath,
		"--output", "text"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect blob inspect opts: %v, got: %v", expected, opts)
	}
}

func TestBlobInspectCommand_SecretsFromEnv(t *testing.T) {
	opts := &blobInspectOpts{}
	expected := &blobInspectOpts{
		signaturePath: "path",
		outputFormat:  cmd.OutputJSON,
	}
	command := inspectCommand(opts)
	if err := command.ParseFlags([]string{
		expected.signaturePath,
		"--output", "json"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect blob inspect opts: %v, got: %v", expected, opts)
	}
}

func TestBlobInspectCommand_MissingArgs(t *testing.T) {
	command := inspectCommand(nil)
	if err := command.ParseFlags(nil); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := command.Args(command, command.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}
