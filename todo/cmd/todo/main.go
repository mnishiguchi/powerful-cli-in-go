package main

import (
	"flag"
	"fmt"
	"os"

	"mnishiguchi.com/todo"
)

// Default filename
var todoFileName = ".todo.json"

/*
## Examples

    # Display the usage
    go run main.go -h

    # List all tasks
    go run main.go -list

    # Add a new task
    go run main.go -task "Go for a walk"
*/
func main() {
	if os.Getenv("TODO_FILENAME") != "" {
		todoFileName = os.Getenv("TODO_FILENAME")
	}

	// Parse command-line flags. See https://pkg.go.dev/flag
	argTask := flag.String("task", "", "A task to be included in the todo list")
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
	case *argTask != "":
		// Add a new task
		list.Add(*argTask)

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
