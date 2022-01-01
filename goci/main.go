package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
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

	// Go relays signals using a channel of type os.Signal.
	// A buffered channel of size 1, which handles at least one signal correctly
	// in case it receives many signals.
	chSignal := make(chan os.Signal, 1)

	// Channels that communicate the status back to the main goroutine.
	chErr := make(chan error)     // notifies potential errors
	chDone := make(chan struct{}) // notifies the loop conclusion

	// The signal.Notify() from the os/signal package relays signals to our
	// chSignal channel. We are only interested in two termination signals:
	// SIGINT and SIGTERM. All other signails will be ignored and not relayed to
	// this channel.
	signal.Notify(chSignal, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// Execute each step looping through the pipeline.
		for _, s := range pipeline {
			successMsg, err := s.execute()
			if err != nil {
				chErr <- err
				return
			}

			_, err = fmt.Fprintln(outWriter, successMsg)
			if err != nil {
				chErr <- err
				return
			}
		}

		close(chDone)
	}()

	// Decide what to do based on communication received in one of the channels.
	for {
		select {
		case receivedSignal := <-chSignal:
			signal.Stop(chSignal) // Stop receiving signals
			return fmt.Errorf("%s: Exiting: %w", receivedSignal, ErrSignal)

		case err := <-chErr:
			return err

		case <-chDone:
			return nil
		}
	}
}
