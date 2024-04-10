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

package command

import (
	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"

	// . "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("notation cert", func() {
	It("show all", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "list").
				MatchKeyWords(
					"STORE TYPE   STORE NAME   CERTIFICATE",
				)
		})
	})

	It("delete all", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "delete", "--all", "--type", "ca", "--store", "e2e", "-y").
				MatchKeyWords(
					"Successfully deleted",
				)
		})
	})

	It("delete a specfic cert", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "delete", "--type", "ca", "--store", "e2e", "e2e.crt", "-y").
				MatchKeyWords(
					"Successfully deleted e2e.crt",
				)
		})
	})

	It("delete a non-exist cert", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.ExpectFailure().Exec("cert", "delete", "--type", "ca", "--store", "e2e", "non-exist.crt", "-y").
				MatchErrKeyWords(
					"failed to delete the certificate file",
				)
		})
	})

	It("show e2e cert", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "list", "--type", "ca", "--store", "e2e").
				MatchKeyWords(
					"e2e.crt",
				)
		})
	})

	It("list", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.Exec("cert", "list").
				MatchKeyWords(
					"STORE TYPE   STORE NAME   CERTIFICATE",
				)
		})
	})
})
