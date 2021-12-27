package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"mnishiguchi.com/todo"
)

// Default filename
var todoFileName = ".todo.json"

/*
## Examples

    cd path/to/cmd/todo

    # Build the executable
    go build

    # Display the usage
    ./todo -h

    # List all tasks
    ./todo -list

    # Add a new task from arguments
    ./todo -add "Go for a walk"

    # Add a new task from STDIN
    ./todo -add
    Study Golang
*/
func main() {
	if os.Getenv("TODO_FILENAME") != "" {
		todoFileName = os.Getenv("TODO_FILENAME")
	}

	// Parse command-line flags. See https://pkg.go.dev/flag
	argAdd := flag.Bool("add", false, "Add a task to the todo list")
	argList := flag.Bool("list", false, "List all tasks")
	argComplete := flag.Int("complete", 0, "Item to be completed")
	flag.Parse()

	// A pointer to an emply todo list
	list := &todo.TodoList{}

	// Read todo items from a file.
	if err := list.Get(todoFileName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Decide what to do based on the number of arguments provided
	switch {
	case *argList:
		// List all todo items
		fmt.Print(list)

	case *argComplete > 0:
		// Complete a given task
		if err := list.Complete(*argComplete); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Save the todo list
		if err := list.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

	case *argAdd:
		task, err := getTask(os.Stdin, flag.Args()...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Add a new task
		list.Add(task)

		// Save the todo list
		if err := list.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

	default:
		// Invalid flag provided
		fmt.Fprintln(os.Stderr, "Invalid option")
		os.Exit(1)
	}
}

// Get the task from either arguments or STDIN.
func getTask(r io.Reader, varargs ...string) (string, error) {
	// If variadic arguments are provided, get the task by joining them.
	if len(varargs) > 0 {
		return strings.Join(varargs, " "), nil
	}

	// Otherwise, read from STDIN.
	s := bufio.NewScanner(r)
	s.Scan()

	if err := s.Err(); err != nil {
		return "", err
	}

	if len(s.Text()) == 0 {
		return "", fmt.Errorf("Task cannot be blank")
	}

	return s.Text(), nil
}
