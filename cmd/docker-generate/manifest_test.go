package main

import (
	"testing"
)

func TestGenerateManifestCmd(t *testing.T) {
	tests := []struct {
		expectedOutput    string
		expectedReference string
		args              []string
		expectedErr       bool
	}{
		{
			expectedOutput:    "abc",
			expectedReference: "def",
			args:              []string{"-o", "abc", "def"},
			expectedErr:       false,
		},
		{
			expectedOutput:    "abc",
			expectedReference: "",
			args:              []string{"-o", "abc"},
			expectedErr:       false,
		},
		{
			expectedOutput:    "",
			expectedReference: "def",
			args:              []string{"def"},
			expectedErr:       false,
		},
		{
			expectedOutput:    "",
			expectedReference: "",
			args:              []string{},
			expectedErr:       false,
		},
		{
			expectedOutput:    "abc",
			expectedReference: "def",
			args:              []string{"def", "--output", "abc"},
			expectedErr:       false,
		},
		{
			args:        []string{"-o", "b", "-n", "x"},
			expectedErr: true,
		},
	}
	for _, test := range tests {
		cmd := generateManifestCommand()
		err := cmd.ParseFlags(test.args)
		if err != nil && !test.expectedErr {
			t.Fatalf("Test failed with error: %v", err)
		}
		if err == nil && test.expectedErr {
			t.Fatalf("Expect test to error but it didn't: %v", test.args)
		}
		if err != nil {
			continue
		}
		if output, _ := cmd.Flags().GetString("output"); output != test.expectedOutput {
			t.Fatalf("Expect output: %v, got: %v", test.expectedOutput, output)
		}
		if arg := cmd.Flags().Arg(0); arg != test.expectedReference {
			t.Fatalf("Expect reference: %v, got: %v", test.expectedReference, arg)
		}
	}

}
