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

// copied and adopted from https://github.com/oras-project/oras with
// modification
/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

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
	workDir string
	timeout time.Duration

	stdin    io.Reader
	exitCode int

	text string

	// env is the environment variables used by the command.
	env map[string]string
}

// Binary returns default execution option for customized binary.
func Binary(binPath string) *ExecOpts {
	return &ExecOpts{
		binPath:  binPath,
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
func (opts *ExecOpts) Exec(args ...string) *Matcher {
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
	description := fmt.Sprintf("\n>> should %s: %s %s >>", opts.text, opts.binPath, strings.Join(args, " "))
	ginkgo.By(description)

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

	// clear ExecOpts state
	opts.Clear()

	return NewMatcher(session)
}

// Clear clears the ExecOpts to get ready for the next execution.
func (opts *ExecOpts) Clear() {
	opts.exitCode = 0
	opts.timeout = DefaultTimeout
	opts.workDir = ""
	opts.stdin = nil
	opts.text = ""
}
