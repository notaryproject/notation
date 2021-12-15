package executor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Executor interface {
	Execute(ctx context.Context, argsf string, a ...interface{}) ([]byte, error)
}

type executor struct {
	cmd     string
	prepend []string
}

func NewExecutor(cmd string, prepend ...string) Executor {
	return &executor{
		cmd:     cmd,
		prepend: prepend,
	}
}

// Execute executes the command with the given arguments.
func (e *executor) Execute(ctx context.Context, argsf string, a ...interface{}) ([]byte, error) {
	substituted := fmt.Sprintf(argsf, a...)
	// "%s %s", "foo", "foo bar" will become ["foo", "foo", "bar"], not [ "foo", "foo bar"]
	// TODO: handle scenarios with space in the arguments
	args := append(e.prepend, strings.Split(substituted, " ")...)
	log.Debugf("Executing: %s %s", e.cmd, strings.Join(args, " "))
	log.Infof("Executing: %s <args>", e.cmd)
	cmd := exec.CommandContext(ctx, e.cmd, args...)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "failed to execute %s %s", e.cmd, strings.Join(args, " "))
	}
	if errb.Len() > 0 {
		return nil, errors.New(errb.String())
	}
	return outb.Bytes(), nil
}
