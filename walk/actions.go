package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Determines whether to filter out or ignore the current path.
func shouldBeExcluded(path string, ext string, minSize int64, info os.FileInfo) bool {
	if info.IsDir() || info.Size() < minSize {
		return true
	}

	if ext != "" && filepath.Ext(path) != ext {
		return true
	}

	return false
}

// Prints out the path of the current file to the specified io.Writer.
func printFilePath(path string, outWriter io.Writer) error {
	_, err := fmt.Fprintln(outWriter, path)
	return err
}
