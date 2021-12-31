package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
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
	}{
		{name: "success",
			projectDir: "./testdata/tool/",
			expected: "go build: SUCCESS\n" +
				"go test: SUCCESS\n" +
				"gofmt: SUCCESS\n" +
				"git push: SUCCESS\n",
			expectedErr:    nil,
			shouldSetupGit: true},
		{name: "fail",
			projectDir:     "./testdata/toolErr/",
			expected:       "",
			expectedErr:    &stepErr{step: "go build"},
			shouldSetupGit: false},
		{name: "failFormat",
			projectDir:     "./testdata/toolFmtErr/",
			expected:       "",
			expectedErr:    &stepErr{step: "go fmt"},
			shouldSetupGit: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// A buffer to capture the output
			var outWriter bytes.Buffer

			if tc.shouldSetupGit {
				cleanupFunc := setupGit(t, tc.projectDir)

				// Ensure that the resources are deleted at the end.
				defer cleanupFunc()
			}

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
