package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

type executer interface {
	execute() (string, error)
}

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

	pipeline := make([]executer, 4)
	// Validate the program's correctness by building the target project without
	// creating an executable file.
	// Running "go build" does not create an executable file when building
	// multiple packages at the same time.
	pipeline[0] = newStep(
		"go build",
		"go",
		"go build: SUCCESS",
		projectDir,
		[]string{"build", ".", "errors"},
	)
	pipeline[1] = newStep(
		"go test",
		"go",
		"go test: SUCCESS",
		projectDir,
		[]string{"test", "-v"},
	)
	// Some programs like "gofmt" exit with a successful return code even when
	// something goes wrong and a message in STDOUT or STDERR provides details
	// about the error condition.
	pipeline[2] = newExceptionStep(
		"go fmt",
		"gofmt",
		"gofmt: SUCCESS",
		projectDir,
		[]string{"-l", "."},
	)
	// As a rule of thumb, when running external commands that can potentially
	// take a long time to complete, it is a good idea to set a timeout.
	pipeline[3] = newTimeoutStep(
		"git push",
		"git",
		"git push: SUCCESS",
		projectDir,
		[]string{"push", "origin", "main"},
		10*time.Second,
	)

	// Execute each step looping through the pipeline.
	for _, s := range pipeline {
		successMsg, err := s.execute()
		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(outWriter, successMsg)
		if err != nil {
			return err
		}
	}

	return nil
}
