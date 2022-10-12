package cmd

import (
	"github.com/spf13/pflag"
)

// SignerFlagOpts cmd opts for using cmd.GetSigner
type SignerFlagOpts struct {
	Key          string
	KeyFile      string
	CertFile     string
	EnvelopeType string
}

// ApplyFlags set flags and their default values for the FlagSet
func (opts *SignerFlagOpts) ApplyFlags(fs *pflag.FlagSet) {
	SetPflagKey(fs, &opts.Key)
	SetPflagKeyFile(fs, &opts.KeyFile)
	SetPflagCertFile(fs, &opts.CertFile)
	SetPflagSignatureFormat(fs, &opts.EnvelopeType)
}
