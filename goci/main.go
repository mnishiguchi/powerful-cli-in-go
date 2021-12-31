package main

import (
	"flag"
	"fmt"
	"io"
	"os"
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

	// _, err := fmt.Fprintln(outWriter, "go build: SUCCESS")
	pipeline := make([]step, 1)
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
