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

package utils

import (
	"fmt"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/gexec"
)

func init() {
	// expand the length limit for the gomega matcher
	format.MaxLength = 1000000
}

// Matcher contains the execution result for matching.
type Matcher struct {
	Session *gexec.Session
	stdout  string
	stderr  string
}

// NewMatcher returns a new Matcher.
func NewMatcher(session *gexec.Session) *Matcher {
	return &Matcher{
		Session: session,
		stdout:  string(session.Out.Contents()),
		stderr:  string(session.Err.Contents()),
	}
}

// MatchContent matches the content with the stdout.
func (m *Matcher) MatchContent(content string) *Matcher {
	Expect(m.stdout).Should(Equal(content))
	return m
}

// MatchErrContent matches the content with stderr.
func (m *Matcher) MatchErrContent(content string) *Matcher {
	Expect(m.stderr).Should(Equal(content))
	return m
}

// MatchKeyWords matches given keywords with the stdout.
func (m *Matcher) MatchKeyWords(keywords ...string) *Matcher {
	matchKeyWords(m.stdout, keywords)
	return m
}

// MatchErrKeyWords matches given keywords with the stderr.
func (m *Matcher) MatchErrKeyWords(keywords ...string) *Matcher {
	matchKeyWords(m.stderr, keywords)
	return m
}

// NoMatchErrKeyWords guarantees that the given keywords do not match with
// the stderr.
func (m *Matcher) NoMatchErrKeyWords(keywords ...string) *Matcher {
	for _, w := range keywords {
		Expect(m.stderr).ShouldNot(ContainSubstring(w))
	}
	return m
}

// MatchErrKeyWords matches given keywords with the stderr.
func matchKeyWords(content string, keywords []string) {
	var missed []string
	lowered := strings.ToLower(content)
	for _, w := range keywords {
		if !strings.Contains(lowered, strings.ToLower(w)) {
			missed = append(missed, w)
		}
	}

	if len(missed) != 0 {
		fmt.Printf("Keywords missed: %v\n", missed)
		panic("failed to match all keywords")
	}
}
