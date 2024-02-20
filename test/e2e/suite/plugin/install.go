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
	"path/filepath"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/onsi/ginkgo/v2"
)

const (
	PluginURL      = "https://github.com/notaryproject/notation-action/raw/e2e-test-plugin/tests/plugin_binaries/notation-e2e-test-plugin_0.1.0_linux_amd64.tar.gz"
	PluginChecksum = "be8d035024d3a96afb4118af32f2e201f126c7254b02f7bcffb3e3149d744fd2"
)

var _ = Describe("notation plugin install", func() {
	It("with missing file or url flag", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", ".").
				MatchErrContent("Error: at least one of the flags in the group [file url] is required\n")
		})
	})

	It("with both file and url flags are set", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", "--file", "--url", ".").
				MatchErrContent("Error: if any flags in the group [file url] are set none of the others can be; [file url] were all set\n")
		})
	})

	It("with missing plugin source", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install").
				MatchErrContent("Error: missing plugin source location\n")
		})
	})

	It("with missing plugin file path", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", "--file").
				MatchErrContent("Error: missing plugin file path\n")
		})
	})

	It("with missing plugin URL", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", "--url").
				MatchErrContent("Error: missing plugin URL\n")
		})
	})

	It("with zip bomb single file exceeds 256 MiB size limit in zip format", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", "--file", filepath.Join(NotationE2EMaliciousPluginArchivePath, "large_file_zip.zip"), "-v").
				MatchErrContent("Error: plugin installation failed: total file size reached the 256 MiB size limit\n")
		})
	})

	It("with zip bomb single file exceeds 256 MiB size limit in tar.gz format", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", "--file", filepath.Join(NotationE2EMaliciousPluginArchivePath, "large_file_tarGz.tar.gz"), "-v").
				MatchErrContent("Error: plugin installation failed: total file size reached the 256 MiB size limit\n")
		})
	})

	It("with zip bomb total file size exceeds 256 MiB size limit", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", "--file", filepath.Join(NotationE2EMaliciousPluginArchivePath, "zip_bomb.zip"), "-v").
				MatchErrContent("Error: plugin installation failed: total file size reached the 256 MiB size limit\n")
		})
	})

	It("with zip slip", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", "--file", filepath.Join(NotationE2EMaliciousPluginArchivePath, "zip_slip.zip"), "-v").
				MatchErrContent("Error: plugin installation failed: file name in zip cannot contain '..', but found \"../../../../../../../../tmp/evil.txt\"\n")
		})
	})

	It("with valid plugin file path", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("plugin", "install", "--file", NotationE2EPluginTarGzPath, "-v").
				MatchContent("Successfully installed plugin e2e-plugin, version 1.0.0\n")
		})
	})

	It("with plugin executable file path", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("plugin", "install", "--file", NotationE2EPluginPath).
				MatchContent("Successfully installed plugin e2e-plugin, version 1.0.0\n")
		})
	})

	It("with plugin already installed", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("plugin", "install", "--file", NotationE2EPluginTarGzPath).
				MatchContent("Successfully installed plugin e2e-plugin, version 1.0.0\n")

			notation.ExpectFailure().Exec("plugin", "install", "--file", NotationE2EPluginTarGzPath).
				MatchErrContent("Error: plugin installation failed: plugin e2e-plugin with version 1.0.0 already exists\n")
		})
	})

	It("with plugin already installed but force install", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("plugin", "install", "--file", NotationE2EPluginTarGzPath, "-v").
				MatchContent("Successfully installed plugin e2e-plugin, version 1.0.0\n")

			notation.Exec("plugin", "install", "--file", NotationE2EPluginTarGzPath, "--force").
				MatchContent("Successfully updated plugin e2e-plugin from version 1.0.0 to 1.0.0\n")
		})
	})

	It("with valid plugin URL", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("plugin", "install", "--url", PluginURL, "--sha256sum", PluginChecksum).
				MatchKeyWords("Successfully installed plugin e2e-test-plugin, version 0.1.0\n")
		})
	})

	It("with valid plugin URL but missing checksum", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", "--url", PluginURL).
				MatchErrContent("Error: installing from URL requires non-empty SHA256 checksum of the plugin source\n")
		})
	})

	It("with valid plugin URL but mismatched SHA-256 checksum", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", "--url", PluginURL, "--sha256sum", "abcd").
				MatchErrContent("Error: plugin installation failed: plugin SHA-256 checksum does not match user input. Expecting abcd\n")
		})
	})

	It("with invalid plugin URL scheme", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", "--url", "http://invalid", "--sha256sum", "abcd").
				MatchErrContent("Error: failed to download plugin from URL: only the HTTPS scheme is supported, but got http\n")
		})
	})

	It("with invalid plugin URL", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", "--url", "https://invalid", "--sha256sum", "abcd").
				MatchErrKeyWords("failed to download plugin from URL https://invalid")
		})
	})
})
