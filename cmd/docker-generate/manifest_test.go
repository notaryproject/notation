package main

import (
	"testing"
)

func TestGenerateManifestCmd(t *testing.T) {
	tests := []struct {
		generateManifestOpts
		args        []string
		expectedErr bool
	}{
		{
			generateManifestOpts{
				reference: "reference",
				output:    "output",
			},
			[]string{"-o", "output", "reference"},
			false,
		},
		{
			generateManifestOpts{
				reference: "",
				output:    "output",
			},
			[]string{"-o", "output"},
			false,
		},
		{
			generateManifestOpts{
				reference: "reference",
				output:    "",
			},
			[]string{"reference"},
			false,
		},
		{
			generateManifestOpts{
				reference: "",
				output:    "",
			},
			[]string{},
			false,
		},
		{
			generateManifestOpts{
				reference: "reference",
				output:    "output",
			},
			[]string{"reference", "--output", "output"},
			false,
		},
		{
			args:        []string{"-o", "output", "-n", "reference"},
			expectedErr: true,
		},
	}
	for _, test := range tests {
		opts := &generateManifestOpts{}
		cmd := generateManifestCommand(opts)
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
		cmd.PreRun(cmd, cmd.Flags().Args())
		if *opts != test.generateManifestOpts {
			t.Fatalf("Expect generate manifest opts: %v, got: %v", test.generateManifestOpts, *opts)
		}
	}

}
