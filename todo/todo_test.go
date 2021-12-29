/* In general, all files in the same directory must belong to the same Go
package. An exeption to this rule is when writing tests. Define the package
name as the original name followed by the "_test" suffix. */
package todo_test

import (
	"os"
	"testing"

	"mnishiguchi.com/todo"
)

func TestAdd(t *testing.T) {
	list := todo.TodoList{}

	taskName := "New task"
	list.Add(taskName)

	if list[0].Task != taskName {
		t.Errorf("Expected %q, got %q instead.", taskName, list[0].Task)
	}
}

func TestComplete(t *testing.T) {
	list := todo.TodoList{}

	taskName := "New task"
	list.Add(taskName)

	if list[0].Task != taskName {
		t.Errorf("Expected %q, got %q instead.", taskName, list[0].Task)
	}

	if list[0].Done {
		t.Errorf("New task should not be completed")
	}

	list.Complete(1)

	if !list[0].Done {
		t.Errorf("New task should be completed")
	}
}

func TestDelete(t *testing.T) {
	list := todo.TodoList{}

	tasks := []string{
		"New task 1",
		"New task 2",
		"New task 3",
	}

	for _, v := range tasks {
		list.Add(v)
	}

	if list[0].Task != tasks[0] {
		t.Errorf("Expected %q, got %q instead.", tasks[0], list[0].Task)
	}

	list.Delete(2)

	if len(list) != 2 {
		t.Errorf("Expected list length %d, got %q instead.", 2, len(list))
	}

	if list[1].Task != tasks[2] {
		t.Errorf("Expected %q, got %q instead.", tasks[2], list[1].Task)
	}
}

func TestSaveGet(t *testing.T) {
	list1 := todo.TodoList{}
	list2 := todo.TodoList{}

	taskName := "New task"
	list1.Add(taskName)

	if list1[0].Task != taskName {
		t.Errorf("Expected %q, got %q instead.", taskName, list1[0].Task)
	}

	// Create a temporary file.
	tmp, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("Error creating temp file: %s", err)
	}
	defer os.Remove(tmp.Name())

	// Save a list to a file.
	if err := list1.Save(tmp.Name()); err != nil {
		t.Fatalf("Error saving list to file: %s", err)
	}

	// Get a list from a file through a new list instance.
	if err := list2.Get(tmp.Name()); err != nil {
		t.Fatalf("Task %q should match %q task.", list1[0].Task, list2[0].Task)
	}
}
