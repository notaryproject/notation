package cert

import "github.com/spf13/cobra"

func Cmd() *cobra.Command {
	command := &cobra.Command{
		Use:     "certificate",
		Aliases: []string{"cert"},
		Short:   "Manage certificates in trust store for signature verification.",
	}

	command.AddCommand(
		certAddCommand(nil),
		certListCommand(nil),
		certShowCommand(nil),
		certDeleteCommand(nil),
		certGenerateTestCommand(nil),
	)

	return command
}
