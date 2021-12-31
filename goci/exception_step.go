package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

/*
This is a variant of the step type that checks the STDOUT for potential error.
Extends the step type. Implements the executer interface.
*/
type exceptionStep struct {
	step // Extends the step type by embedding the step type.
}

func (s exceptionStep) execute() (string, error) {
	cmd := exec.Command(s.executable, s.args...)

	var outBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	cmd.Dir = s.targetProjectDir

	// Run the command and check for potential errors.
	if err := cmd.Run(); err != nil {
		return "", &stepErr{
			step:  s.name,
			msg:   "failed to execute",
			cause: err,
		}
	}

	// Verify the size of the output buffer. If it contains anything, that
	// means some code in the project does not match the correct format.
	// Currently this is not generic enough, coupled with "gofmt".
	if outBuffer.Len() > 0 {
		return "", &stepErr{
			step:  s.name,
			msg:   fmt.Sprintf("invalid format: %s", outBuffer.String()),
			cause: nil,
		}
	}

	return s.successMsg, nil
}
