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

package plugin

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	notationplugin "github.com/notaryproject/notation/cmd/notation/internal/plugin"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/spf13/cobra"
)

func TestInstall(t *testing.T) {
	t.Run("invalid plugin source url", func(t *testing.T) {
		opts := &pluginInstallOpts{
			pluginSourceType: notationplugin.PluginSourceTypeURL,
			inputChecksum:    "dummy",
			pluginSource:     "http://[::1]/%",
		}
		expectedErrMsg := `failed to parse plugin download URL http://[::1]/% with error: parse "http://[::1]/%": invalid URL escape "%"`
		err := install(&cobra.Command{}, opts)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected error %s, but got %s", expectedErrMsg, err)
		}
	})

	t.Run("unknown plugin source type", func(t *testing.T) {
		opts := &pluginInstallOpts{
			pluginSourceType: -1,
		}
		expectedErrMsg := `plugin installation failed: unknown plugin source type`
		err := install(&cobra.Command{}, opts)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected error %s, but got %s", expectedErrMsg, err)
		}
	})
}

func TestInstallPlugin(t *testing.T) {
	ctx := context.Background()
	t.Run("input path does not exist", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping test on Windows")
		}
		expectedErrMsg := `stat invalid: no such file or directory`
		err := installPlugin(ctx, "invalid", "", false)
		if err == nil || err.Error() != expectedErrMsg {
			t.Fatalf("expected error %s, but got %s", expectedErrMsg, err)
		}
	})

	t.Run("failed to get file type", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping test on Windows")
		}

		tempDir := t.TempDir()
		data := []byte("data")
		filename := filepath.Join(tempDir, "a", "file.txt")
		if err := osutil.WriteFile(filename, data); err != nil {
			t.Fatal(err)
		}
		err := os.Chmod(tempDir, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := os.Chmod(tempDir, 0700)
			if err != nil {
				t.Fatal(err)
			}
		}()

		expectedErrMsg := `permission denied`
		err = installPlugin(ctx, filename, "", false)
		if err == nil || !strings.Contains(err.Error(), expectedErrMsg) {
			t.Fatalf("expected permission denied error, but got %s", err)
		}
	})
}
