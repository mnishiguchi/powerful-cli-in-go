package main

import (
	"context"
	"os/exec"
	"time"
)

/*
This is a variant of the step type that times out in case the executed program
is hanging.

Extends the step type. Implements the executer interface.
*/
type timeoutStep struct {
	step
	timeout time.Duration
}

// A constructor of the timeoutStep type.
// Go does not have formal constructors like other OO languages but it is a good
// practice to ensure callers instantiate types correctly.
func newTimeoutStep(
	name,
	executable,
	successMsg,
	targetProjectDir string,
	args []string,
	timeout time.Duration,
) timeoutStep {
	s := timeoutStep{}

	s.step = step{
		name:             name,
		executable:       executable,
		successMsg:       successMsg,
		targetProjectDir: targetProjectDir,
		args:             args,
	}

	s.timeout = timeout
	if s.timeout == 0 {
		// Default to 30 seconds
		s.timeout = 30 * time.Second
	}

	return s
}

// When testing, we override the original function with the mock function.
// For a more robust approach, we can pass the function we want to override as a
// parameter or use an interface.
var cmdWithContext = exec.CommandContext

func (s timeoutStep) execute() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	// Free up the resources when context is no longer needed.
	defer cancel()

	cmd := cmdWithContext(ctx, s.executable, s.args...)
	cmd.Dir = s.targetProjectDir

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", &stepErr{
				step:  s.name,
				msg:   "failed time out",
				cause: context.DeadlineExceeded,
			}
		}

		return "", &stepErr{
			step:  s.name,
			msg:   "failed to execute",
			cause: err,
		}
	}

	return s.successMsg, nil
}
