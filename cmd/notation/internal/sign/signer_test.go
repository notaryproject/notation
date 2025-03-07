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

package sign

import (
	"context"
	"runtime"
	"sync"
	"testing"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/signer"
	"github.com/notaryproject/notation/cmd/notation/internal/flag"
	"github.com/notaryproject/notation/internal/config"
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
	opts := &flag.SignerFlagOpts{
		KeyID:      "testKeyId",
		PluginName: "testPlugin",
	}

	_, err := GetSigner(ctx, opts)
	if err != nil {
		t.Fatalf("expected nil error, but got %s", err)
	}

	_, err = GetSigner(ctx, opts)
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
	opts := &flag.SignerFlagOpts{
		Key: "test",
	}

	_, err := GetSigner(ctx, opts)
	if err != nil {
		t.Fatalf("expected nil error, but got %s", err)
	}

	_, err = GetSigner(ctx, opts)
	if err != nil {
		t.Fatalf("expected nil error, but got %s", err)
	}
}

func TestGetFailed(t *testing.T) {
	ctx := context.Background()
	opts := &flag.SignerFlagOpts{}

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
}

func TestGetSignerFailed(t *testing.T) {
	ctx := context.Background()

	defer func(oldLibexeDir, oldConfigDir string) {
		dir.UserLibexecDir = oldLibexeDir
		dir.UserConfigDir = oldConfigDir
	}(dir.UserLibexecDir, dir.UserConfigDir)

	t.Run("get failed", func(t *testing.T) {
		opts := &flag.SignerFlagOpts{}
		dir.UserLibexecDir = "./testdata/plugins"
		dir.UserConfigDir = "./testdata/invalid_signingkeys"
		_, err := GetSigner(ctx, opts)
		if err == nil {
			t.Fatal("GetSigner should return an error")
		}
	})

	t.Run("invalid plugin name in opts", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping test on Windows")
		}

		dir.UserLibexecDir = "./testdata/plugins"
		dir.UserConfigDir = "./testdata/invalid_signingkeys"
		opts := &flag.SignerFlagOpts{
			KeyID:      "test",
			PluginName: "invalid",
		}
		expectedErrMsg := `plugin executable file is either not found or inaccessible: stat testdata/plugins/plugins/invalid/notation-invalid: no such file or directory`
		_, err := GetSigner(ctx, opts)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected %s, but got %s", expectedErrMsg, err)
		}
	})

	t.Run("failed to resolve key", func(t *testing.T) {
		dir.UserConfigDir = "./testdata/no_default_key_signingkeys"
		expectedErrMsg := `default signing key not set. Please set default signing key or specify a key name`
		_, err := GetSigner(ctx, &flag.SignerFlagOpts{})
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected %s, but got %s", expectedErrMsg, err)
		}
	})

	t.Run("keypath not specified", func(t *testing.T) {
		dir.UserConfigDir = "./testdata/invalid_signingkeys"
		expectedErrMsg := `key path not specified`
		opts := &flag.SignerFlagOpts{
			Key: "invalid",
		}
		_, err := GetSigner(ctx, opts)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected %s, but got %s", expectedErrMsg, err)
		}
	})

	t.Run("key not found", func(t *testing.T) {
		dir.UserConfigDir = "./testdata/valid_signingkeys"
		expectedErrMsg := `signing key not found`
		opts := &flag.SignerFlagOpts{
			Key: "test2",
		}
		_, err := GetSigner(ctx, opts)
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
		opts := &flag.SignerFlagOpts{
			Key: "invalidExternal",
		}
		_, err := GetSigner(ctx, opts)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected %s, but got %s", expectedErrMsg, err)
		}
	})

	t.Run("empty key", func(t *testing.T) {
		dir.UserConfigDir = "./testdata/invalid_signingkeys"
		expectedErrMsg := `unsupported key, either provide a local key and certificate file paths, or a key name in config.json, check https://notaryproject.dev/docs/user-guides/how-to/notation-config-file/ for details`
		opts := &flag.SignerFlagOpts{
			Key: "empty",
		}
		_, err := GetSigner(ctx, opts)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected %s, but got %s", expectedErrMsg, err)
		}
	})
}

func TestResolveKey(t *testing.T) {
	defer func(oldDir string) {
		dir.UserConfigDir = oldDir
		config.LoadConfigOnce = sync.OnceValues(config.LoadConfig)
	}(dir.UserConfigDir)

	t.Run("valid test key", func(t *testing.T) {
		dir.UserConfigDir = "./testdata/valid_signingkeys"
		keySuite, err := resolveKey("test")
		if err != nil {
			t.Fatal(err)
		}
		if keySuite.Name != "test" {
			t.Error("key name is not correct.")
		}
	})

	t.Run("key name is empty (using default key)", func(t *testing.T) {
		dir.UserConfigDir = "./testdata/valid_signingkeys"
		keySuite, err := resolveKey("")
		if err != nil {
			t.Fatal(err)
		}
		if keySuite.Name != "test" {
			t.Error("key name is not correct.")
		}
	})

	t.Run("signingkeys.json without read permission", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping test on Windows")
		}
		dir.UserConfigDir = "./testdata/empty_signingkey"

		_, err := resolveKey("")
		expectedErrMsg := "default signing key not set. Please set default signing key or specify a key name"
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected %s, but got %s", expectedErrMsg, err)
		}
	})
}
