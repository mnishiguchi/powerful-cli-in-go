package main

import (
	"flag"
	"fmt"
	"io"
	"os"
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

	pipeline := make([]executer, 3)
	// Validate the program's correctness by building the target project without
	// creating an executable file.
	// Running "go build" does not create an executable file when building
	// multiple packages at the same time.
	pipeline[0] = step{
		name:             "go build",
		executable:       "go",
		successMsg:       "go build: SUCCESS",
		targetProjectDir: projectDir,
		args:             []string{"build", ".", "errors"},
	}
	pipeline[1] = step{
		name:             "go test",
		executable:       "go",
		successMsg:       "go test: SUCCESS",
		targetProjectDir: projectDir,
		args:             []string{"test", "-v"},
	}
	// Some programs like "gofmt" exit with a successful return code even when
	// something goes wrong and a message in STDOUT or STDERR provides details
	// about the error condition.
	pipeline[2] = exceptionStep{
		step: step{
			name:             "go fmt",
			executable:       "gofmt",
			successMsg:       "gofmt: SUCCESS",
			targetProjectDir: projectDir,
			args:             []string{"-l", "."},
		},
	}

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
