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
	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/onsi/ginkgo/v2"
)

const (
	PluginURL      = "https://github.com/notaryproject/notation-action/raw/e2e-test-plugin/tests/plugin_binaries/notation-e2e-test-plugin_0.1.0_linux_amd64.tar.gz"
	PluginCheckSum = "be8d035024d3a96afb4118af32f2e201f126c7254b02f7bcffb3e3149d744fd2"
)

var _ = Describe("notation plugin install", func() {
	It("with valid plugin file path", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("plugin", "install", "--file", NotationE2EPluginTarGzPath, "-v").
				MatchContent("Succussefully installed plugin e2e-plugin, version 1.0.0\n")
		})
	})

	It("with invalid plugin file type", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", "--file", NotationE2EPluginPath).
				MatchErrContent("Error: failed to install the plugin: invalid file format. Only support .tar.gz and .zip\n")
		})
	})

	It("with plugin already installed", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("plugin", "install", "--file", NotationE2EPluginTarGzPath).
				MatchContent("Succussefully installed plugin e2e-plugin, version 1.0.0\n")

			notation.ExpectFailure().Exec("plugin", "install", "--file", NotationE2EPluginTarGzPath).
				MatchErrContent("Error: failed to install the plugin: e2e-plugin with version 1.0.0 already exists\n")
		})
	})

	It("with plugin already installed but force install", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("plugin", "install", "--file", NotationE2EPluginTarGzPath, "-v").
				MatchContent("Succussefully installed plugin e2e-plugin, version 1.0.0\n")

			notation.Exec("plugin", "install", "--file", NotationE2EPluginTarGzPath, "--force").
				MatchContent("Succussefully installed plugin e2e-plugin, updated the version from 1.0.0 to 1.0.0\n")
		})
	})

	It("with valid plugin URL", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("plugin", "install", "--url", PluginURL, "--sha256sum", PluginCheckSum).
				MatchContent("Succussefully installed plugin e2e-test-plugin, version 0.1.0\n")
		})
	})

	It("with valid plugin URL but missing checksum", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", "--url", PluginURL).
				MatchErrContent("Error: install from URL requires non-empty SHA256 checksum of the plugin source\n")
		})
	})

	It("with invalid plugin URL scheme", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", "--url", "http://invalid", "--sha256sum", "abcd").
				MatchErrContent("Error: failed to install from URL: \"http://invalid\" scheme is not HTTPS\n")
		})
	})

	It("with invalid plugin URL", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", "--url", "https://invalid", "--sha256sum", "abcd").
				MatchErrKeyWords("failed to download plugin from URL https://invalid")
		})
	})
})
