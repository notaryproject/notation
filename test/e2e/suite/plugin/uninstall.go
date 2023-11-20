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

var _ = Describe("notation plugin uninstall", func() {
	It("with valid plugin name", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			vhost.SetOption(AddPlugin(NotationE2EPluginPath))
			notation.Exec("plugin", "uninstall", "--yes", "e2e-plugin").
				MatchContent("Successfully uninstalled plugin e2e-plugin\n")
		})
	})

	It("with plugin does not exist", func() {
		Host(nil, func(notation *utils.ExecOpts, _ *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("plugin", "uninstall", "--yes", "non-exist").
				MatchErrContent("Error: unable to find plugin non-exist.\nTo view a list of installed plugins, use `notation plugin list`\n")
		})
	})

})
