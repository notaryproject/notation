package utils

import (
	"fmt"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

type Matcher struct {
	Session *gexec.Session
	stdout  string
	stderr  string
}

func NewMatcher(session *gexec.Session) *Matcher {
	return &Matcher{
		Session: session,
		stdout:  string(session.Out.Contents()),
		stderr:  string(session.Err.Contents()),
	}
}

func (m *Matcher) MatchContent(content string) *Matcher {
	Expect(m.stdout).Should(Equal(content))
	return m
}

func (m *Matcher) MatchErrContent(content string) *Matcher {
	Expect(m.stderr).Should(Equal(content))
	return m
}

func (m *Matcher) MatchKeyWords(keywords ...string) *Matcher {
	matchKeyWords(m.stdout, keywords)
	return m
}

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
