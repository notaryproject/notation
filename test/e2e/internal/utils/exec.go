package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/notaryproject/notation/test/e2e/internal/utils/match"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const (
	DefaultTimeout = 10 * time.Second
	// If the command hasn't exited yet, ginkgo session ExitCode is -1
	notResponding = -1
)

// ExecOpts is an option used to execute a command.
type ExecOpts struct {
	binPath string
	args    []string
	workDir string
	timeout time.Duration

	stdin    io.Reader
	stdout   []match.Matcher
	stderr   []match.Matcher
	exitCode int

	text string

	// env is the environment variables used by the command.
	env map[string]string
}

// Binary returns default execution option for customized binary.
func Binary(binPath string, args ...string) *ExecOpts {
	return &ExecOpts{
		binPath:  binPath,
		args:     args,
		timeout:  DefaultTimeout,
		exitCode: 0,
		env:      make(map[string]string),
	}
}

// ExpectFailure sets failure exit code checking for the execution.
func (opts *ExecOpts) ExpectFailure() *ExecOpts {
	// set to 1 but only check if it's positive
	opts.exitCode = 1
	return opts
}

// ExpectBlocking consistently check if the execution is blocked.
func (opts *ExecOpts) ExpectBlocking() *ExecOpts {
	opts.exitCode = notResponding
	return opts
}

// WithTimeOut sets timeout for the execution.
func (opts *ExecOpts) WithTimeOut(timeout time.Duration) *ExecOpts {
	opts.timeout = timeout
	return opts
}

// WithDescription sets description text for the execution.
func (opts *ExecOpts) WithDescription(text string) *ExecOpts {
	opts.text = text
	return opts
}

// WithWorkDir sets working directory for the execution.
func (opts *ExecOpts) WithWorkDir(path string) *ExecOpts {
	opts.workDir = path
	return opts
}

// WithInput redirects stdin to r for the execution.
func (opts *ExecOpts) WithInput(r io.Reader) *ExecOpts {
	opts.stdin = r
	return opts
}

// MatchKeyWords adds keywords matching to stdout.
func (opts *ExecOpts) MatchKeyWords(keywords ...string) *ExecOpts {
	opts.stdout = append(opts.stdout, match.NewKeywordMatcher(keywords))
	return opts
}

// MatchErrKeyWords adds keywords matching to stderr.
func (opts *ExecOpts) MatchErrKeyWords(keywords ...string) *ExecOpts {
	opts.stderr = append(opts.stderr, match.NewKeywordMatcher(keywords))
	return opts
}

// MatchContent adds full content matching to the execution.
func (opts *ExecOpts) MatchContent(content string) *ExecOpts {
	if opts.exitCode == 0 {
		opts.stdout = append(opts.stdout, match.NewContentMatcher(content, false))
	} else {
		opts.stderr = append(opts.stderr, match.NewContentMatcher(content, false))
	}
	return opts
}

// MatchTrimedContent adds trimmed content matching to the execution.
func (opts *ExecOpts) MatchTrimmedContent(content string) *ExecOpts {
	if opts.exitCode == 0 {
		opts.stdout = append(opts.stdout, match.NewContentMatcher(content, true))
	} else {
		opts.stderr = append(opts.stderr, match.NewContentMatcher(content, true))
	}
	return opts
}

// Clear clears the ExecOpts to ready for the next execution.
func (opts *ExecOpts) Clear() {
	opts.args = nil
	opts.exitCode = 0
	opts.timeout = DefaultTimeout
	opts.workDir = ""
	opts.stdin = nil
	opts.stdout = nil
	opts.stderr = nil
	opts.text = ""
}

// WithEnv update the environment variables.
func (opts *ExecOpts) WithEnv(env map[string]string) *ExecOpts {
	if env == nil {
		return opts
	}
	if opts.env == nil {
		opts.env = make(map[string]string)
	}
	for key, value := range env {
		opts.env[key] = value
	}
	return opts
}

// Exec run the execution based on opts.
func (opts *ExecOpts) Exec(args ...string) *gexec.Session {
	if opts == nil {
		// this should be a code error but can only be caught during runtime
		panic("Nil option for command execution")
	}

	if opts.text == "" {
		// set default description text
		switch opts.exitCode {
		case notResponding:
			opts.text = "block"
		case 0:
			opts.text = "pass"
		default:
			opts.text = "fail"
		}
	}
	description := fmt.Sprintf("\n>> should %s: %s %s >>", opts.text, opts.binPath, strings.Join(opts.args, " "))
	ginkgo.By(description)

	// overwrite the args
	if len(args) == 0 {
		args = opts.args
	}

	var cmd *exec.Cmd
	cmd = exec.Command(opts.binPath, args...)

	// set environment variables
	cmd.Env = append(cmd.Env, os.Environ()...)
	for key, val := range opts.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%v=%v", key, val))
	}

	// set stdin
	cmd.Stdin = opts.stdin
	if opts.workDir != "" {
		// switch working directory
		wd, err := os.Getwd()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(os.Chdir(opts.workDir)).ShouldNot(HaveOccurred())
		defer os.Chdir(wd)
	}
	fmt.Println(description)
	session, err := gexec.Start(cmd, os.Stdout, os.Stderr)
	Expect(err).ShouldNot(HaveOccurred())
	if opts.exitCode == notResponding {
		Consistently(session.ExitCode).WithTimeout(opts.timeout).Should(Equal(notResponding))
		session.Kill()
	} else {
		exitCode := session.Wait(opts.timeout).ExitCode()
		Expect(opts.exitCode == 0).To(Equal(exitCode == 0))
	}

	// matching result
	for _, s := range opts.stdout {
		s.Match(session.Out)
	}
	for _, s := range opts.stderr {
		s.Match(session.Err)
	}

	// clear ExecOpts state
	opts.Clear()

	return session
}
