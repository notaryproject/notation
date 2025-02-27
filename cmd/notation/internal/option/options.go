// Copyright The Notary Project Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package option

import (
	"context"
	"os"

	"github.com/notaryproject/notation-go/log"
	"github.com/notaryproject/notation/cmd/notation/internal/constant"
	"github.com/notaryproject/notation/internal/trace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"oras.land/oras-go/v2/registry/remote/auth"
	credentialstrace "oras.land/oras-go/v2/registry/remote/credentials/trace"
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

// InitializeLogger sets up the logger based on common options.
func (opts *LoggingFlagOpts) InitializeLogger(ctx context.Context) context.Context {
	if opts.Debug {
		ctx = trace.WithLoggerLevel(ctx, logrus.DebugLevel)
	} else if opts.Verbose {
		ctx = trace.WithLoggerLevel(ctx, logrus.InfoLevel)
	} else {
		return ctx
	}
	return withExecutableTrace(ctx)
}

// withExecutableTrace adds tracing for credential helper executables.
func withExecutableTrace(ctx context.Context) context.Context {
	logger := log.GetLogger(ctx)
	ctx = credentialstrace.WithExecutableTrace(ctx, &credentialstrace.ExecutableTrace{
		ExecuteStart: func(executableName, action string) {
			logger.Debugf("started executing credential helper program %s with action %s", executableName, action)
		},
		ExecuteDone: func(executableName, action string, err error) {
			if err != nil {
				logger.Errorf("finished executing credential helper program %s with action %s and got error %w", executableName, action, err)
			} else {
				logger.Debugf("successfully finished executing credential helper program %s with action %s", executableName, action)
			}
		},
	})
	return ctx
}

// SecureFlagOpts defines flags for registry credentials and security
type SecureFlagOpts struct {
	Username         string
	Password         string
	InsecureRegistry bool
}

// ApplyFlags set flags and their default values for the FlagSet
func (opts *SecureFlagOpts) ApplyFlags(fs *pflag.FlagSet) {
	setflagUsername(fs, &opts.Username)
	setFlagPassword(fs, &opts.Password)
	setFlagInsecureRegistry(fs, &opts.InsecureRegistry)
	opts.Username = os.Getenv(constant.UsernameEnv)
	opts.Password = os.Getenv(constant.PasswordEnv)
}

// Credential returns an auth.Credential from opts.Username and opts.Password.
func (opts *SecureFlagOpts) Credential() auth.Credential {
	if opts.Username == "" {
		return auth.Credential{
			RefreshToken: opts.Password,
		}
	}
	return auth.Credential{
		Username: opts.Username,
		Password: opts.Password,
	}
}
