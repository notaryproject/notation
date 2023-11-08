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
	It("with invalid plugin file type", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", NotationE2EPluginPath).
				MatchErrContent("Error: failed to install the plugin, invalid file format. Only support .tar.gz and .zip\n")
		})
	})

	It("with valid plugin file path", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("plugin", "install", NotationE2EPluginTarGzPath, "-v").
				MatchContent("Succussefully installed plugin e2e-plugin, version 1.0.0\n")
		})
	})

	It("with plugin already installed", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", NotationE2EPluginTarGzPath).
				MatchErrContent("plugin e2e-plugin already installed\n")
		})
	})

	It("with plugin already installed but force install", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("plugin", "install", NotationE2EPluginTarGzPath, "--force").
				MatchContent("Succussefully installed plugin e2e-plugin, version 1.0.0\n")
		})
	})

	It("with valid plugin URL but missing checksum", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", PluginURL).
				MatchErrContent("install from URL requires non-empty SHA256 checksum of the plugin source")
		})
	})

	It("with invalid plugin URL scheme", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", "http://invalid", "--checksum", "abcd").
				MatchErrContent("input plugin URL has to be in scheme HTTPS, but got http")
		})
	})

	It("with invalid plugin URL", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "install", "https://invalid", "--checksum", "abcd").
				MatchErrKeyWords("failed to download plugin from URL https://invalid")
		})
	})

	It("with valid plugin URL", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("plugin", "install", PluginURL, "--checksum", PluginCheckSum).
				MatchContent("Succussefully installed plugin e2e-test-plugin, version 0.1.0\n")
		})
	})
})
