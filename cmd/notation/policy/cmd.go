package policy

import "github.com/spf13/cobra"

func Cmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "policy [command]",
		Short: "[Preview] Manage trust policy configuration",
		Long:  "[Preview] Manage trust policy configuration for signature verification.",
	}

	command.AddCommand(
		showCmd(),
		importCmd(),
	)

	return command
}
