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

package cmd

import (
	"context"
	"runtime"
	"testing"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/signer"
)

func TestGenericSignerImpl(t *testing.T) {
	g := &signer.GenericSigner{}
	if _, ok := interface{}(g).(notation.Signer); !ok {
		t.Fatal("GenericSigner does not implement notation.Signer")
	}

	if _, ok := interface{}(g).(notation.BlobSigner); !ok {
		t.Fatal("GenericSigner does not implement notation.BlobSigner")
	}
}

func TestPluginSignerImpl(t *testing.T) {
	p := &signer.PluginSigner{}
	if _, ok := interface{}(p).(notation.Signer); !ok {
		t.Fatal("PluginSigner does not implement notation.Signer")
	}

	if _, ok := interface{}(p).(notation.BlobSigner); !ok {
		t.Fatal("PluginSigner does not implement notation.BlobSigner")
	}
}

func TestGetSignerFromOpts(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on Windows")
	}

	defer func(oldLibexeDir string) {
		dir.UserLibexecDir = oldLibexeDir
	}(dir.UserLibexecDir)

	dir.UserLibexecDir = "./testdata/plugins"
	ctx := context.Background()
	opts := &SignerFlagOpts{
		KeyID:      "testKeyId",
		PluginName: "testPlugin",
	}

	_, err := GetSigner(ctx, opts)
	if err != nil {
		t.Fatalf("expected nil error, but got %s", err)
	}

	_, err = GetBlobSigner(ctx, opts)
	if err != nil {
		t.Fatalf("expected nil error, but got %s", err)
	}
}

func TestGetSignerFromConfig(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on Windows")
	}

	defer func(oldLibexeDir, oldConfigDir string) {
		dir.UserLibexecDir = oldLibexeDir
		dir.UserConfigDir = oldConfigDir
	}(dir.UserLibexecDir, dir.UserConfigDir)

	dir.UserLibexecDir = "./testdata/plugins"
	dir.UserConfigDir = "./testdata/valid_signingkeys"
	ctx := context.Background()
	opts := &SignerFlagOpts{
		Key: "test",
	}

	_, err := GetSigner(ctx, opts)
	if err != nil {
		t.Fatalf("expected nil error, but got %s", err)
	}

	_, err = GetBlobSigner(ctx, opts)
	if err != nil {
		t.Fatalf("expected nil error, but got %s", err)
	}
}

func TestGetFailed(t *testing.T) {
	ctx := context.Background()
	opts := &SignerFlagOpts{}

	defer func(oldLibexeDir, oldConfigDir string) {
		dir.UserLibexecDir = oldLibexeDir
		dir.UserConfigDir = oldConfigDir
	}(dir.UserLibexecDir, dir.UserConfigDir)

	dir.UserLibexecDir = "./testdata/plugins"
	dir.UserConfigDir = "./testdata/invalid_signingkeys"
	_, err := GetSigner(ctx, opts)
	if err == nil {
		t.Fatal("GetSigner should return an error")
	}

	_, err = GetBlobSigner(ctx, opts)
	if err == nil {
		t.Fatal("GetBlobSigner should return an error")
	}
}

func TestSignerCore(t *testing.T) {
	ctx := context.Background()

	defer func(oldLibexeDir, oldConfigDir string) {
		dir.UserLibexecDir = oldLibexeDir
		dir.UserConfigDir = oldConfigDir
	}(dir.UserLibexecDir, dir.UserConfigDir)

	t.Run("invalid plugin name in opts", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping test on Windows")
		}

		dir.UserLibexecDir = "./testdata/plugins"
		dir.UserConfigDir = "./testdata/invalid_signingkeys"
		opts := &SignerFlagOpts{
			KeyID:      "test",
			PluginName: "invalid",
		}
		expectedErrMsg := `plugin executable file is either not found or inaccessible: stat testdata/plugins/plugins/invalid/notation-invalid: no such file or directory`
		_, err := signerCore(ctx, opts)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected %s, but got %s", expectedErrMsg, err)
		}
	})

	t.Run("failed to resolve key", func(t *testing.T) {
		dir.UserConfigDir = "./testdata/valid_signingkeys"
		expectedErrMsg := `default signing key not set. Please set default signing key or specify a key name`
		_, err := signerCore(ctx, &SignerFlagOpts{})
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected %s, but got %s", expectedErrMsg, err)
		}
	})

	t.Run("keypath not specified", func(t *testing.T) {
		dir.UserConfigDir = "./testdata/invalid_signingkeys"
		expectedErrMsg := `key path not specified`
		opts := &SignerFlagOpts{
			Key: "invalid",
		}
		_, err := signerCore(ctx, opts)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected %s, but got %s", expectedErrMsg, err)
		}
	})

	t.Run("key not found", func(t *testing.T) {
		dir.UserConfigDir = "./testdata/valid_signingkeys"
		expectedErrMsg := `signing key not found`
		opts := &SignerFlagOpts{
			Key: "test2",
		}
		_, err := signerCore(ctx, opts)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected %s, but got %s", expectedErrMsg, err)
		}
	})

	t.Run("invalid plugin name in signingkeys", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping test on Windows")
		}

		dir.UserLibexecDir = "./testdata/plugins"
		dir.UserConfigDir = "./testdata/invalid_signingkeys"
		expectedErrMsg := `plugin executable file is either not found or inaccessible: stat testdata/plugins/plugins/invalid/notation-invalid: no such file or directory`
		opts := &SignerFlagOpts{
			Key: "invalidExternal",
		}
		_, err := signerCore(ctx, opts)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected %s, but got %s", expectedErrMsg, err)
		}
	})

	t.Run("empty key", func(t *testing.T) {
		dir.UserConfigDir = "./testdata/invalid_signingkeys"
		expectedErrMsg := `unsupported key, either provide a local key and certificate file paths, or a key name in config.json, check https://notaryproject.dev/docs/user-guides/how-to/notation-config-file/ for details`
		opts := &SignerFlagOpts{
			Key: "empty",
		}
		_, err := signerCore(ctx, opts)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected %s, but got %s", expectedErrMsg, err)
		}
	})
}
