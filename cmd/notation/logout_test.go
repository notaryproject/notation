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
	"context"
	"errors"
	"testing"
)

func TestLogoutCommand_BasicArgs(t *testing.T) {
	opts := &logoutOpts{}
	cmd := logoutCommand(opts)
	expected := &logoutOpts{
		server: "server",
	}
	if err := cmd.ParseFlags([]string{
		expected.server,
	}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err != nil {
		t.Fatalf("Parse Args failed: %v", err)
	}
	if *opts != *expected {
		t.Fatalf("Expect logout opts: %v, got: %v", expected, opts)
	}
}

func TestLogOutCommand_MissingArgs(t *testing.T) {
	cmd := logoutCommand(nil)
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if err := cmd.Args(cmd, cmd.Flags().Args()); err == nil {
		t.Fatal("Parse Args expected error, but ok")
	}
}

// The errors returned by runLogout have to wrap the underlying error so that
// callers can inspect it with errors.Is and errors.As.
func TestRunLogout_WrapsErrors(t *testing.T) {
	t.Run("invalid credentials store config", func(t *testing.T) {
		setTestConfigDir(t, "{")

		err := runLogout(context.Background(), &logoutOpts{server: "registry.example.com"})
		if err == nil {
			t.Fatal("runLogout() expected an error, but got nil")
		}
		if errors.Unwrap(err) == nil {
			t.Fatalf("runLogout() error does not wrap the underlying error: %v", err)
		}
	})

	t.Run("credentials not found", func(t *testing.T) {
		setTestConfigDir(t, "{}")

		err := runLogout(context.Background(), &logoutOpts{server: "registry.example.com"})
		if err == nil {
			t.Skip("logging out of an unknown registry does not fail on this platform")
		}
		if errors.Unwrap(err) == nil {
			t.Fatalf("runLogout() error does not wrap the underlying error: %v", err)
		}
	})
}
