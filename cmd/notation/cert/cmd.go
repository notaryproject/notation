package cert

import (
	"github.com/notaryproject/notation/cmd/notation/internal/experimental"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	command := &cobra.Command{
		Use:     "certificate",
		Aliases: []string{"cert"},
		Short:   "Manage certificates in trust store",
		Long:    "Manage certificates in trust store for signature verification.",
	}

	command.AddCommand(
		certAddCommand(nil),
		certListCommand(nil),
		certShowCommand(nil),
		certDeleteCommand(nil),
		certGenerateTestCommand(nil),
	)
	if !experimental.IsDisabled() {
		command.AddCommand(
			certCleanupTestCommand(nil),
		)
	}

	return command
}
