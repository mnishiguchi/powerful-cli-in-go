package main_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
		fmt.Fprintf(os.Stderr, "Cannnot build tool %s: %s", binName, err)
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

	taskName := "test task number 1"

	// Add a new task
	t.Run("AddNewTask", func(t *testing.T) {
		cmd := exec.Command(cmdPath, strings.Split(taskName, " ")...)

		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	})

	// List all tasks
	t.Run("ListTasks", func(t *testing.T) {
		cmd := exec.Command(cmdPath)
		cmdOutput, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}
		expected := taskName + "\n"

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
