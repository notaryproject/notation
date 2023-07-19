package cmd

import (
	"context"

	"github.com/notaryproject/notation-go/log"
	"github.com/notaryproject/notation/internal/trace"
	executableTrace "github.com/oras-project/oras-credentials-go/trace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// SignerFlagOpts cmd opts for using cmd.GetSigner
type SignerFlagOpts struct {
	Key             string
	SignatureFormat string
	KeyID           string
	PluginName      string
}

// ApplyFlags set flags and their default values for the FlagSet
func (opts *SignerFlagOpts) ApplyFlagsToCommand(command *cobra.Command) {
	fs := command.Flags()
	SetPflagKey(fs, &opts.Key)
	SetPflagSignatureFormat(fs, &opts.SignatureFormat)
	SetPflagID(fs, &opts.KeyID)
	SetPflagPlugin(fs, &opts.PluginName)
	command.MarkFlagsRequiredTogether("id", "plugin")
	command.MarkFlagsMutuallyExclusive("key", "id")
	command.MarkFlagsMutuallyExclusive("key", "plugin")
}

// LoggingFlagOpts option struct.
type LoggingFlagOpts struct {
	Debug   bool
	Verbose bool
}

// ApplyFlags applies flags to a command flag set.
func (opts *LoggingFlagOpts) ApplyFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&opts.Debug, "debug", "d", false, "debug mode")
	fs.BoolVarP(&opts.Verbose, "verbose", "v", false, "verbose mode")
}

// SetLoggerLevel sets up the logger based on common options.
func (opts *LoggingFlagOpts) SetLoggerLevel(ctx context.Context) context.Context {
	logger := log.GetLogger(ctx)
	ctx = executableTrace.WithExecutableTrace(ctx, &executableTrace.ExecutableTrace{
		ExecuteStart: func(executableName, action string) {
			logger.Info("started executing credential helper program %s with action %s", executableName, action)
		},
		ExecuteDone: func(executableName, action string, err error) {
			logger.Info("finished executing credential helper program %s with action %s and erro %v", executableName, action, err)
		},
	})
	if opts.Debug {
		return trace.WithLoggerLevel(ctx, logrus.DebugLevel)
	} else if opts.Verbose {
		return trace.WithLoggerLevel(ctx, logrus.InfoLevel)
	}
	return ctx
}
