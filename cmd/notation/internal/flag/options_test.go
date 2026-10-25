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

package flag

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"oras.land/oras-go/v2/registry/remote/auth"
	credentialstrace "oras.land/oras-go/v2/registry/remote/credentials/trace"

	"github.com/notaryproject/notation-go/log"
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
			want: auth.Credential{},
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

func TestWithExecutableTrace_ExecuteDoneLogsError(t *testing.T) {
	var out bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&out)
	logger.SetLevel(logrus.DebugLevel)

	ctx := withExecutableTrace(log.WithLogger(context.Background(), logger))
	credentialstrace.ContextExecutableTrace(ctx).ExecuteDone("docker-credential-test", "get", errors.New("boom"))

	got := out.String()
	if !strings.Contains(got, "and got error boom") {
		t.Errorf("the logged error is missing the error message, got %q", got)
	}
	if strings.Contains(got, "%!") {
		t.Errorf("the logged error contains a formatting error, got %q", got)
	}
}
