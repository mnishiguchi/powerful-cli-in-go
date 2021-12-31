package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func main() {
	projectDir := flag.String("p", "", "Go Project directory")
	flag.Parse()

	if err := run(*projectDir, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// The main logic of this program.
func run(projectDir string, outWriter io.Writer) error {
	if projectDir == "" {
		return fmt.Errorf("Project directory is required: %w", ErrValidation)
	}

	// Validate the program's correctness by building the target project without
	// creating an executable file.

	// Running "go build" does not create an executable file when building
	// multiple packages at the same time. So we add one extra package from the
	// Go standard library to the argument list.
	args := []string{"build", ".", "errors"}
	cmd := exec.Command("go", args...)
	// set the working directory for the external command execution.
	cmd.Dir = projectDir
	if err := cmd.Run(); err != nil {
		// return fmt.Errorf("'go build' failed: %s", err)
		return &stepErr{
			step:  "go build",
			msg:   "go build failed",
			cause: err,
		}
	}

	_, err := fmt.Fprintln(outWriter, "go build: SUCCESS")
	return err
}
