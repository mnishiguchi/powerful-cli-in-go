package todo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type TodoItem struct {
	Task        string
	Done        bool
	CreatedAt   time.Time
	CompletedAt time.Time
}

type TodoList []TodoItem

// Creates a new TODO item and appends it to the list.
func (l *TodoList) Add(task string) {
	item := TodoItem{
		Task:        task,
		Done:        false,
		CreatedAt:   time.Now(),
		CompletedAt: time.Time{},
	}
	*l = append(*l, item)
}

// Marks an item as completed.
func (l *TodoList) Complete(index int) error {
	if index <= 0 || index > len(*l) {
		return fmt.Errorf("TodoItem %d does not exist", index)
	}

	// Adjust index for 0-based index.
	(*l)[index-1].Done = true
	(*l)[index-1].CompletedAt = time.Now()

	return nil
}

// Removes an item from the list.
func (l *TodoList) Delete(index int) error {
	if index <= 0 || index > len(*l) {
		return fmt.Errorf("TodoItem %d does not exist", index)
	}

	// Adjust index for 0-based index.
	*l = append((*l)[:index-1], (*l)[index:]...)

	return nil
}

// Encodes the list as JSON and saves it using the provided file name.
func (l *TodoList) Save(filename string) error {
	jsonifiedList, err := json.Marshal(l)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, jsonifiedList, 0644)
}

// Opens the provided file name, decodes the JSON data and parses it into a list.
func (l *TodoList) Get(filename string) error {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	if len(file) == 0 {
		return nil
	}

	return json.Unmarshal(file, l)
}