package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type config struct {
	ext  string // extension to filter out
	size int64  // min file size
	list bool   // list files
}

// ## Examples
//
//    ❯ go run . -root tmp/testdir
//
//    # Filter by min file size in bytes
//    ❯ go run . -root tmp/testdir -size 10
//
//    # Filter by file extension
//    ❯ go run . -root tmp/testdir -ext .log
//
func main() {
	// Define and parse command-line flags
	argRootDir := flag.String("root", ".", "Root directory to start")
	argExt := flag.String("ext", "", "File extension to filter out")
	argSize := flag.Int64("size", 0, "Minimum file size")
	argList := flag.Bool("list", false, "List files only")
	flag.Parse()

	cfg := config{
		ext:  *argExt,
		size: *argSize,
		list: *argList,
	}

	if err := run(*argRootDir, os.Stdout, cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Coordinates operations:
//  * descend into the provided root directory
//  * find all the files and sub-directories under the root directory
//  * filter the list based on the specified options
func run(rootDir string, outWriter io.Writer, cfg config) error {
	return filepath.Walk(
		rootDir,
		// an annonymous function
		func(path string, info os.FileInfo, err error) error {
			// Check if it was successful to walk to this file or directory.
			if err != nil {
				return err
			}

			// Check if the current file or directory should be filtered out.
			if shouldBeExcluded(path, cfg.ext, cfg.size, info) {
				return nil
			}

			// If the list option was explicitly set, do nothing else.
			if cfg.list {
				return printFilePath(path, outWriter)
			}

			// Do the "list" action if nothing else was set.
			return printFilePath(path, outWriter)
		})
}
