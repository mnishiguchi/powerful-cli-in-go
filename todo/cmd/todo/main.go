package main

import (
	"fmt"
	"os"
	"strings"

	"mnishiguchi.com/todo"
)

// Hardcode the file name for now
const todoFileName = ".todo.json"

/*
## Examples

    # List all tasks
    go run main.go

    # Add a new task
    go run main.go Go for a walk

*/
func main() {
	// A pointer to an emply todo list
	list := &todo.TodoList{}

	// Read todo items from a file.
	if err := list.Get(todoFileName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Decide what to do based on the number of arguments provided
	switch {
	// For no extra argument other than the program name
	case len(os.Args) == 1:
		// Print all the todo items
		for _, item := range *list {
			fmt.Println(item.Task)
		}
	// For extra arguments
	default:
		// Concatenate all the extra arguments with a space and add the task
		task := strings.Join(os.Args[1:], " ")
		list.Add(task)

		// Save the new list
		if err := list.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
