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

package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/spf13/pflag"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func TestSecureFlagOpts_Credential(t *testing.T) {
	tests := []struct {
		name string
		opts *SecureFlagOpts
		want auth.Credential
	}{
		{
			name: "Username and password",
			opts: &SecureFlagOpts{
				Username: "username",
				Password: "password",
			},
			want: auth.Credential{
				Username: "username",
				Password: "password",
			},
		},
		{
			name: "Username only",
			opts: &SecureFlagOpts{
				Username: "username",
			},
			want: auth.Credential{
				Username: "username",
			},
		},
		{
			name: "Password only",
			opts: &SecureFlagOpts{
				Password: "token",
			},
			want: auth.Credential{
				RefreshToken: "token",
			},
		},
		{
			name: "Empty username and password",
			opts: &SecureFlagOpts{
				Username: "",
				Password: "",
			},
			want: auth.EmptyCredential,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.opts.Credential(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SecureFlagOpts.Credential() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCredentialsUnset(t *testing.T) {
	const notationUsername = "NOTATION_USERNAME"
	const notationPassword = "NOTATION_PASSWORD"
	// Set environment variables for testing
	os.Setenv(notationUsername, "testuser")
	os.Setenv(notationPassword, "testpassword")

	secureFlagOpts := &SecureFlagOpts{}
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	secureFlagOpts.ApplyFlags(fs)

	// check credentials environment variables are unset
	if os.Getenv(notationUsername) != "" {
		t.Errorf("expected %s to be unset", notationUsername)
	}

	if os.Getenv(notationPassword) != "" {
		t.Errorf("expected %s to be unset", notationPassword)
	}
}
