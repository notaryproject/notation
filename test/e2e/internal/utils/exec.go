package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

// ExecOpts is an option used to execute a command.
type ExecOpts struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	// Env is the environment variables used by the command.
	Env map[string]string
}

// CommandOpts is an option used by CommandGroup to execute a batch of notation commands.
// TODO: how to read data in a container in a convenient way
type CommandOpts struct {
	ExecOpts
	Description string
	// Binary is an executable file. The default value will be notation binary if not provided.
	Binary string
	// Args is arguments to execute the binary.
	Args       []string
	ShouldFail bool
	// Checker is an user-provided function used to validate the result of a Command.
	Checker func(CommandOpts, *gexec.Session)
}

// CommandGroup contains a batch of e2e notation command to be executed.
type CommandGroup []CommandOpts

// CommandGroupOpts is option used to create a CommandGroup.
type CommandGroupOpts func(g CommandGroup)

// WithAuth sets up auth info for the CommandGroup.
func WithAuth(username, password string) CommandGroupOpts {
	return func(g CommandGroup) {
		for i, c := range g {
			g[i].ExecOpts = c.WithAuth(username, password)
		}
	}
}

// WithUserDir sets up user config and cache directory for the CommandGroup.
func WithUserDir(dir string) CommandGroupOpts {
	return func(g CommandGroup) {
		for i, c := range g {
			g[i].ExecOpts = c.WithUserDir(dir)
		}
	}
}

// NewCommandGroup creates a CommandsGroup from base.
func NewCommandGroup(base CommandGroup, opts ...CommandGroupOpts) CommandGroup {
	for _, opt := range opts {
		opt(base)
	}
	return base
}

// WithAuth creates an ExecOpts with auth info setted.(By setting $NOTATION_USERNAME and $NOTATION_PASSWORD)
func (opts ExecOpts) WithAuth(username, password string) ExecOpts {
	if opts.Env == nil {
		opts.Env = make(map[string]string)
	}
	opts.Env["NOTATION_USERNAME"] = username
	opts.Env["NOTATION_PASSWORD"] = password
	return opts
}

// WithUserDir creates an ExecOpts with user config and cache directory setted(By setting $XDG_CONFIG_HOME and $XDG_CACHE_HOME).
func (opts ExecOpts) WithUserDir(dir string) ExecOpts {
	if opts.Env == nil {
		opts.Env = make(map[string]string)
	}
	configDir, cacheDir := filepath.Join(dir, "config"), filepath.Join(dir, "cache")
	os.MkdirAll(configDir, os.ModePerm)
	os.MkdirAll(cacheDir, os.ModePerm)
	opts.Env["XDG_CONFIG_HOME"] = configDir
	opts.Env["XDG_CACHE_HOME"] = cacheDir
	return opts
}

// Exec execuates a binary with args and opts.
func Exec(binary string, opts ExecOpts, args ...string) (*gexec.Session, error) {
	cmd := exec.Command(binary, args...)
	cmd.Env = append(cmd.Env, os.Environ()...)
	for key, val := range opts.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%v=%v", key, val))
	}
	if opts.Stdin != nil {
		cmd.Stdin = opts.Stdin
	}
	session, err := gexec.Start(cmd, opts.Stdout, opts.Stderr)
	if err != nil {
		return nil, err
	}
	session = session.Wait("30s")
	return session, nil
}

func description(text string, binary string, args []string) string {
	return fmt.Sprintf("%s: %s %s", text, binary, strings.Join(args, " "))
}

// batchExec executes a batch of notation commands. If containerID is provided, it will execute all commands in a container.
func batchExec(text string, commands *CommandGroup, containerID string) {
	for _, command := range *commands {
		if command.ExecOpts.Stdout == nil {
			command.ExecOpts.Stdout = GinkgoWriter
		}
		if command.ExecOpts.Stderr == nil {
			command.ExecOpts.Stderr = GinkgoWriter
		}

		name := command.Binary
		args := command.Args
		if containerID != "" {
			name = "docker"
			binary := command.Binary
			// by default call notation
			if command.Binary == "" {
				binary = "notation"
			}
			args = append([]string{"exec", containerID, binary}, command.Args...)
		} else {
			// by default call notation
			if command.Binary == "" {
				name = NotationBinaryPath
			}
		}

		By(description(text, name, args))
		session, err := Exec(name, command.ExecOpts, args...)
		Expect(err).ShouldNot(HaveOccurred())
		var exitCode int
		if command.ShouldFail {
			Expect(session.ExitCode()).NotTo(Equal(exitCode))
		} else {
			Expect(session.ExitCode()).To(Equal(exitCode))
		}
		if command.Checker != nil {
			command.Checker(command, session)
		}
	}
}

// ExecCommandGroup executes every notation command in a single spec on the host machine.
// If binary is not provided in a single command, command will execuate notation with args.
// Otherwise command will execute the given binary with args.
func ExecCommandGroup(text string, commands *CommandGroup) {
	It(fmt.Sprintf("[%s]", text), func() {
		batchExec(text, commands, "")
	})
}

// ExecCommandGroupWithSysEnv executes commands in a container.
// User and system config/cache will be isolated from the host machine.
// This function is typically used to test system level config.
// Environment variables won't be set in this case.
// It will first create  a container if containerID is not provided.
// If binary is not provided in a single command, command will execuate notation with args.
// Otherwise command will execute the given binary with args.
func ExecCommandGroupInContainer(text string, commands *CommandGroup) {
	It(fmt.Sprintf("[%s]", text), func() {
		containerID, _, err := SetUpContainer()
		Expect(err).ShouldNot(HaveOccurred())
		batchExec(text, commands, containerID)
	})
}
