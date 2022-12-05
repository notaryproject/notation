package cmd

import (
	"context"

	"github.com/notaryproject/notation/internal/trace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// SignerFlagOpts cmd opts for using cmd.GetSigner
type SignerFlagOpts struct {
	Key             string
	SignatureFormat string
}

// ApplyFlags set flags and their default values for the FlagSet
func (opts *SignerFlagOpts) ApplyFlags(fs *pflag.FlagSet) {
	SetPflagKey(fs, &opts.Key)
	SetPflagSignatureFormat(fs, &opts.SignatureFormat)
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
	if opts.Debug {
		return trace.WithLoggerLevel(ctx, logrus.DebugLevel)
	} else if opts.Verbose {
		return trace.WithLoggerLevel(ctx, logrus.InfoLevel)
	}
	return ctx
}
