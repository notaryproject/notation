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
	"os"
	"runtime"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation plugin list", func() {
	It("with empty result", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("plugin", "list").
				MatchContent("NAME   DESCRIPTION   VERSION   CAPABILITIES   ERROR   \n")
		})
	})

	It("with e2e-plugin installed", func() {
		Host(Opts(AddPlugin(NotationE2EPluginPath)), func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("plugin", "list").
				MatchKeyWords("NAME", "e2e-plugin").
				MatchKeyWords("DESCRIPTION", "The e2e-plugin is a Notation compatible plugin for Notation E2E test").
				MatchKeyWords("VERSION", "1.0.0").
				MatchKeyWords("CAPABILITIES", "[SIGNATURE_VERIFIER.TRUSTED_IDENTITY SIGNATURE_VERIFIER.REVOCATION_CHECK]").
				MatchKeyWords("ERROR", "<nil>")
		})
	})

	It("missing plugin binary", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			// create azure-kv plugin directory
			pluginDir := vhost.AbsolutePath(NotationDirName, "plugins", "azure-kv")
			if err := os.MkdirAll(pluginDir, os.ModePerm); err != nil {
				Fail(err.Error())
			}

			notation.Exec("plugin", "list").
				MatchKeyWords("azure-kv").
				MatchKeyWords("not found")
		})
	})

	It("with invalid binary file", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			// create azure-kv plugin directory
			pluginDir := vhost.AbsolutePath(NotationDirName, "plugins", "azure-kv")
			if err := os.MkdirAll(pluginDir, os.ModePerm); err != nil {
				Fail(err.Error())
			}

			// create invalid plugin binary
			invalidPluginBinary := vhost.AbsolutePath(NotationDirName, "plugins", "azure-kv", "notation-azure-kv")
			if runtime.GOOS == "windows" {
				invalidPluginBinary += ".exe"
			}
			if err := os.WriteFile(invalidPluginBinary, []byte("invalid"), 0755); err != nil {
				Fail(err.Error())
			}

			notation.Exec("plugin", "list").
				MatchKeyWords("azure-kv").
				MatchKeyWords("not executable")
		})
	})
})
