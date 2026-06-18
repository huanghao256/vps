package runner

import (
	"context"
	"errors"
	"os/exec"
	"slices"
	"time"
)

type CommandRunner struct {
	allowed []string
	timeout time.Duration
}

func NewCommandRunner(allowed []string, timeout time.Duration) CommandRunner {
	return CommandRunner{
		allowed: slices.Clone(allowed),
		timeout: timeout,
	}
}

func (r CommandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	if !slices.Contains(r.allowed, name) {
		return nil, errors.New("command is not allowlisted: " + name)
	}

	if r.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}
