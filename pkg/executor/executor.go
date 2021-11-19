package executor

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Interface interface {
	Execute(ctx context.Context, argsf string, a ...interface{}) ([]byte, error)
}

type executor struct {
	cmd     string
	prepend []string
}

var _ Interface = &executor{}

func NewExecutor(cmd string, prepend ...string) Interface {
	return &executor{
		cmd:     cmd,
		prepend: prepend,
	}
}

func (e *executor) Execute(ctx context.Context, argsf string, a ...interface{}) ([]byte, error) {
	substituted := fmt.Sprintf(argsf, a...)
	args := append(e.prepend, strings.Split(substituted, " ")...)
	log.Infof("Executing: %s %s", e.cmd, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, e.cmd, args...)
	out, err := cmd.CombinedOutput()
	log.Infof("Output: %s", out)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute %s %s", e.cmd, strings.Join(args, " "))
	}
	return out, nil
}
