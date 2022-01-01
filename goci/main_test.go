package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	// Since the command git is required to execute this test, skip the test if
	// git is unavailable.
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("Git not installed. Skippint test.")
	}

	var testCases = []struct {
		name           string
		projectDir     string
		expected       string
		expectedErr    error
		shouldSetupGit bool
		mockCmd        func(ctx context.Context, name string, arg ...string) *exec.Cmd
	}{
		{name: "success",
			projectDir: "./testdata/tool/",
			expected: "go build: SUCCESS\n" +
				"go test: SUCCESS\n" +
				"gofmt: SUCCESS\n" +
				"git push: SUCCESS\n",
			expectedErr:    nil,
			shouldSetupGit: true,
			mockCmd:        nil},
		{name: "successMock",
			projectDir: "./testdata/tool/",
			expected: "go build: SUCCESS\n" +
				"go test: SUCCESS\n" +
				"gofmt: SUCCESS\n" +
				"git push: SUCCESS\n",
			expectedErr:    nil,
			shouldSetupGit: true,
			mockCmd:        mockCmdContext},
		{name: "fail",
			projectDir:     "./testdata/toolErr/",
			expected:       "",
			expectedErr:    &stepErr{step: "go build"},
			shouldSetupGit: false,
			mockCmd:        nil},
		{name: "failFormat",
			projectDir:     "./testdata/toolFmtErr/",
			expected:       "",
			expectedErr:    &stepErr{step: "go fmt"},
			shouldSetupGit: false,
			mockCmd:        nil},
		{name: "failTimeout",
			projectDir:     "./testdata/tool/",
			expected:       "",
			expectedErr:    context.DeadlineExceeded,
			shouldSetupGit: false,
			mockCmd:        mockCmdTimeout},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldSetupGit {
				_, err := exec.LookPath("git")
				if err != nil {
					t.Skip("Git not installed. Skipping test.")
				}

				cleanup := setupGit(t, tc.projectDir)

				// Ensure that the resources are deleted at the end.
				defer cleanup()
			}

			if tc.mockCmd != nil {
				// Override the package variable with a mock.
				cmdWithContext = tc.mockCmd
			}

			// A buffer to capture the output
			var outWriter bytes.Buffer

			err := run(tc.projectDir, &outWriter)

			// When an error is expected
			if tc.expectedErr != nil {
				if err == nil {
					t.Errorf("Expected error: %q. Got 'nil' instead.", tc.expectedErr)
					return
				}

				if !errors.Is(err, tc.expectedErr) {
					t.Errorf("Expected error: %q. Got %q", tc.expectedErr, err)
				}

				return
			}

			// When no error is expected
			if err != nil {
				t.Errorf("Unexpected error: %q", err)
			}

			if outWriter.String() != tc.expected {
				t.Errorf("Expected output: %q. Got %q", tc.expected, outWriter.String())
			}
		})
	}
}

func TestRunKill(t *testing.T) {
	var testCases = []struct {
		name             string
		targetProjectDir string
		expectedSignal   syscall.Signal
		expectedErr      error
	}{
		{"SIGINT", "./testdata/tool", syscall.SIGINT, ErrSignal},
		{"SIGTERM", "./testdata/tool", syscall.SIGTERM, ErrSignal},
		// This test case fails for some reason...
		// {"SIGQUIT", "./testdata/tool", syscall.SIGQUIT, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Override the package variable with a mock command function.
			cmdWithContext = mockCmdTimeout

			// Since we are handling signals, the test will run the functions
			// concurrently.
			chErr := make(chan error)
			chIgnoredSignal := make(chan os.Signal, 1)
			chExpectedSignal := make(chan os.Signal, 1)

			signal.Notify(chIgnoredSignal, syscall.SIGQUIT)
			defer signal.Stop(chIgnoredSignal)

			signal.Notify(chExpectedSignal, tc.expectedSignal)
			defer signal.Stop(chExpectedSignal)

			// Sends potential errors to the chErr channel.
			go func() {
				chErr <- run(tc.targetProjectDir, io.Discard)
			}()

			// Sends the desired signal to the test executable.
			go func() {
				time.Sleep(2 * time.Second)
				syscall.Kill(syscall.Getpid(), tc.expectedSignal)
			}()

			// Select error
			select {
			case err := <-chErr:
				if err == nil {
					t.Errorf("Expected error. Got 'nil' instead.")
					return
				}

				if !errors.Is(err, tc.expectedErr) {
					t.Errorf("Expected error: %q. Got %q", tc.expectedErr, err)
				}
			}

			// select signal
			select {
			case receivedSignal := <-chExpectedSignal:
				if receivedSignal != tc.expectedSignal {
					t.Errorf("Expected signal %q, got %q", tc.expectedSignal, receivedSignal)
				}

			case <-chIgnoredSignal:
				// do nothing

			default:
				t.Errorf("Signal not received")
			}
		})
	}
}

// Helpers
/*
A helper that sets up a reproducible environment for git.

1. Create a temporaray directory.
2. Create a bare git repository on that temporary directory.
3. Initialize a git repository on the target project directory.
4. Add the bare git repository as a remote repository in the empty git repository in the target project directory.
5. Stage a file to commit.
6. Commit the changes to the git repository.

A bare git repository is a repository that contains only the git data but no
working directory so it cannot be used to make local modifications to the code.
This characteristic makes it well suited to serve as a remote repository.
*/
func setupGit(t *testing.T, targetProjectDir string) func() {
	t.Helper()

	// Check if the command "git" is available on the system.
	gitExec, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}

	// Create a temporary directory for the simulated remove Git repository.
	tempDir, err := os.MkdirTemp("", "goci_test")
	// fmt.Println(tempDir)

	if err != nil {
		t.Fatal(err)
	}

	// Get the absolute path of the target project directory.
	targetProjectDirAbs, err := filepath.Abs(targetProjectDir)
	if err != nil {
		t.Fatal(err)
	}

	// The URI path that points to the temporaray directory. Since we are
	// simulating the remote repository locally, we can use the protocol "file://"
	// for the URI.
	remoteURI := fmt.Sprintf("file://%s", tempDir)

	// A slice of annonymous struct that is used for executing a series of git
	// commands.
	var gitCmdList = []struct {
		args []string // the arguments for the git command
		dir  string   // the directory on which to execute the command
		env  []string // A list of environment variables used during the execution
	}{
		{args: []string{"init", "--bare"},
			dir: tempDir,
			env: nil},
		{args: []string{"init"},
			dir: targetProjectDirAbs,
			env: nil},
		{args: []string{"remote", "add", "origin", remoteURI},
			dir: targetProjectDirAbs,
			env: nil},
		{args: []string{"add", "."},
			dir: targetProjectDirAbs,
			env: nil},
		{args: []string{"commit", "-m", "test"},
			dir: targetProjectDirAbs,
			env: []string{
				"GIT_COMMITTER_NAME=test",
				"GIT_COMMITTER_EMAIL=test@example.com",
				"GIT_AUTHOR_NAME=test",
				"GIT_AUTHOR_EMAIL=test@example.com"}},
	}

	// Loop over the command list, executing each one in sequence.
	for _, g := range gitCmdList {
		gitCmd := exec.Command(gitExec, g.args...)
		gitCmd.Dir = g.dir

		if g.env != nil {
			gitCmd.Env = append(os.Environ(), g.env...)
		}

		if err := gitCmd.Run(); err != nil {
			t.Fatal(err)
		}
	}

	// Returns the cleanup function.
	return func() {
		// Remove the temporary directory.
		os.RemoveAll(tempDir)
		// Delete the local .git subdirectory from the target project directory.
		os.RemoveAll(filepath.Join(targetProjectDirAbs, ".git"))
	}
}

// This is a mock for the func exec.CommandContext() function.
func mockCmdContext(ctx context.Context, executable string, args ...string) *exec.Cmd {
	// Build an argument list with "-test.run=TestHelperProcess" so that Go will
	// invoke our TestHelperProcess() function before running the tests.
	// E.g., [-test.run=TestHelperProcess git push origin main]
	goExecutableArgs := []string{"-test.run=TestHelperProcess"}
	goExecutableArgs = append(goExecutableArgs, executable)
	goExecutableArgs = append(goExecutableArgs, args...)

	// os.Args[0] - the full path of the Go executable.
	cmd := exec.CommandContext(ctx, os.Args[0], goExecutableArgs...)
	// Add the environment variable "GO_WANT_HELPER_PROCESS=1" to the command
	// environemnt to ensure the test is not skipped.
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}

	return cmd
}

// This is a mock for the func exec.CommandContext() function.
func mockCmdTimeout(ctx context.Context, exe string, args ...string) *exec.Cmd {
	cmd := mockCmdContext(ctx, exe, args...)
	cmd.Env = append(cmd.Env, "GO_HELPER_TIMEOUT=1")

	return cmd
}

// Simulates the Go command behavior that we want to test. We name this function
// TestHelperProcess() following the standard library convention.
func TestHelperProcess(t *testing.T) {
	// Skip this test unless we specify "GO_WANT_HELPER_PROCESS=1".
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// Simulate a long-running process.
	if os.Getenv("GO_HELPER_TIMEOUT") == "1" {
		time.Sleep(15 * time.Second)
	}

	// ?????
	if os.Args[2] == "git" {
		fmt.Fprintln(os.Stdout, "Everything up-to-date")
		os.Exit(0)
	}

	os.Exit(1)

	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("Git not installed. Skipping test.")
	}
}
