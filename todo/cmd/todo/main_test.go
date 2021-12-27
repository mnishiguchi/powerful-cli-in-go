package main_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

var (
	binName  = "todo"
	fileName = ".todo.json"
)

// Executes extra setup before our tests
func TestMain(m *testing.M) {
	fmt.Println("Building tool...")

	// When this test is running on Windows OS, the executable file extension
	// would be ".exe".
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	// Build the executable binary for our CLI tool.
	build := exec.Command("go", "build", "-o", binName)

	// Make sure that the executable works.
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot build tool %s: %s", binName, err)
		os.Exit(1)
	}

	fmt.Println("Running tests...")
	result := m.Run()

	fmt.Println("Cleaning up...")

	// Delete files that are used in our tests.
	os.Remove(binName)
	os.Remove(fileName)

	os.Exit(result)
}

func TestTodoCLI(t *testing.T) {
	cmdPath, err := findExecutable()
	if err != nil {
		t.Fatal(err)
	}

	taskName1 := "test task number 1"

	// Add a new task from arguments
	t.Run("AddNewTaskFromArguments", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-add", taskName1)

		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	})

	taskName2 := "test task number 2"

	// Add a new task from STDIN
	t.Run("AddNewTaskFromSTDIN", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-add")
		cmdStdin, err := cmd.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		io.WriteString(cmdStdin, taskName2)
		cmdStdin.Close()

		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	})

	// List all tasks
	t.Run("ListTasks", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-list")
		cmdOutput, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}
		expected := fmt.Sprintf("[ ] 1: %s\n[ ] 2: %s\n", taskName1, taskName2)

		if expected != string(cmdOutput) {
			t.Errorf("Expected %q, got %q instead\n", expected, string(cmdOutput))
		}
	})
}

// Find the executable that is compiled in TestMain()
func findExecutable() (string, error) {
	currentDir, err := os.Getwd()
	return filepath.Join(currentDir, binName), err
}
