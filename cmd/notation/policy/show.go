package policy

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/spf13/cobra"
)

type showOpts struct {
}

func showCmd() *cobra.Command {
	var opts showOpts
	command := &cobra.Command{
		Use:   "show [flags]",
		Short: "[Preview] Show trust policy configuration",
		Long: `[Preview] Show trust policy configuration.

** This command is in preview and under development. **

Example - Show current trust policy configuration:
  notation policy show

Example - Save current trust policy configuration to a file:
  notation policy show > my_policy.json
`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(cmd, opts)
		},
	}
	return command
}

func runShow(command *cobra.Command, opts showOpts) error {
	// get policy file path
	policyPath, err := dir.ConfigFS().SysPath(dir.PathTrustPolicy)
	if err != nil {
		return fmt.Errorf("failed to obtain path of trust policy configuration file: %w", err)
	}

	// core process
	policyJSON, err := os.ReadFile(policyPath)
	if err != nil {
		return fmt.Errorf("failed to load trust policy configuration, you may import one via `notation policy import <path-to-policy.json>`: %w", err)
	}
	var doc trustpolicy.Document
	if err = json.Unmarshal(policyJSON, &doc); err == nil {
		err = doc.Validate()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "Existing trust policy configuration is invalid, you may update or create a new one via `notation policy import <path-to-policy.json>`\n")
		// not returning to show the invalid policy configuration
	}

	// show policy content
	_, err = os.Stdout.Write(policyJSON)
	return err
}
