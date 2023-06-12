package policy

import "github.com/spf13/cobra"

func Cmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "policy [command]",
		Short: "Manage trust policy configuration",
		Long:  "Manage trust policy configuration for signature verification.",
	}

	command.AddCommand(
		showCmd(),
		importCmd(),
	)

	return command
}
