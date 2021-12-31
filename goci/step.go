package main

import "os/exec"

// Implements the executer interface.
type step struct {
	name             string
	executable       string
	args             []string
	successMsg       string
	targetProjectDir string
}

// Instantiates a new step.
// Go does not have formal constructors like other OO languages but it is a good
// practice to ensure callers instantiate types correctly.
func newStep(
	name,
	executable,
	successMsg,
	targetProjectDir string,
	args []string,
) step {
	return step{
		name:             name,
		executable:       executable,
		successMsg:       successMsg,
		targetProjectDir: targetProjectDir,
		args:             args,
	}
}

func (s step) execute() (string, error) {
	cmd := exec.Command(s.executable, s.args...)
	// The working directory for the external command execution.
	cmd.Dir = s.targetProjectDir

	if err := cmd.Run(); err != nil {
		return "", &stepErr{
			step:  s.name,
			msg:   "failed to execute",
			cause: err,
		}
	}

	return s.successMsg, nil
}
