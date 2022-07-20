package cmd

import (
	"github.com/spf13/pflag"
)

// SignerFlagOpts cmd opts for using cmd.GetSigner
type SignerFlagOpts struct {
	Key      string
	KeyFile  string
	CertFile string
}

func (opts *SignerFlagOpts) ApplyFlag(fs *pflag.FlagSet) {
	SetPflagKey(fs, &opts.Key)
	SetPflagKeyFile(fs, &opts.KeyFile)
	SetPflagCertFile(fs, &opts.CertFile)
}
