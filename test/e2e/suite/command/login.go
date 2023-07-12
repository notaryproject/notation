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
	"fmt"

	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/notaryproject/notation/test/e2e/suite/common"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("notation login", func() {
	BeforeEach(func() {
		Skip("The login tests require setting up credential helper running in host and it is not available in Github runner. Issue to remove this skip: https://github.com/notaryproject/notation/issues/587")
	})
	It("should sign an image after successfully logging in the registry by prompt with a correct credential", func() {
		Host(TestLoginOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.WithInput(gbytes.BufferWithBytes([]byte(fmt.Sprintf("%s\n%s\n", TestRegistry.Username, TestRegistry.Password)))).
				Exec("login", artifact.Host).
				MatchKeyWords(LoginSuccessfully)
			notation.Exec("sign", artifact.ReferenceWithDigest()).
				MatchKeyWords(SignSuccessfully)
			notation.Exec("logout", artifact.Host).
				MatchKeyWords(LogoutSuccessfully)
		})
	})

	It("should fail to sign an image after failing to log in the registry with a wrong credential", func() {
		Host(TestLoginOptions(), func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost) {
			notation.WithInput(gbytes.BufferWithBytes([]byte(fmt.Sprintf("%s\n%s\n", "invalidUser", "invalidPassword")))).
				ExpectFailure().
				Exec("login", artifact.Host).
				MatchErrKeyWords("unauthorized")
			notation.ExpectFailure().
				Exec("sign", artifact.ReferenceWithDigest()).
				MatchErrKeyWords("credential required for basic auth")
		})
	})
})
