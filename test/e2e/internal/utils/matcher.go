package utils

import (
	"fmt"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

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
