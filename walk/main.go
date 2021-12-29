package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type config struct {
	ext            string // extension to filter out
	size           int64  // min file size
	isListAction   bool   // list files
	isDeleteAction bool   // delete files
	logWriter      io.Writer
	archiveDir     string
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
//    # Delete matched files, logging to STDOUT
//    ❯ go run . -root tmp/testdir -ext .log -delete
//
//    # Delete matched files, logging to a specified file
//    ❯ go run . -root tmp/testdir -ext .log -delete -log log/deleted_files.log
//
func main() {
	// Define and parse command-line flags
	argRootDir := flag.String("root", ".", "Root directory to start")
	argLogFile := flag.String("log", "", "Log deletes to this file")
	// Action options
	argListAction := flag.Bool("list", false, "List files only")
	argDeleteAction := flag.Bool("delete", false, "Delete files")
	argArchiveDir := flag.String("archive", "", "Archive directory")
	// Filter options
	argFilterByExt := flag.String("ext", "", "File extension to filter out")
	argFilterBySize := flag.Int64("size", 0, "Minimum file size")
	flag.Parse()

	var (
		f   = os.Stdout
		err error
	)

	if *argLogFile != "" {
		f, err = os.OpenFile(*argLogFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
	}

	cfg := config{
		ext:            *argFilterByExt,
		size:           *argFilterBySize,
		isListAction:   *argListAction,
		isDeleteAction: *argDeleteAction,
		logWriter:      f,
		archiveDir:     *argArchiveDir,
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
	logger := log.New(cfg.logWriter, "DELETED FILE: ", log.LstdFlags)

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
			if cfg.isListAction {
				return printFilePath(path, outWriter)
			}

			if cfg.archiveDir != "" {
				if err := archiveFile(cfg.archiveDir, rootDir, path); err != nil {
					return err
				}
			}

			if cfg.isDeleteAction {
				return deleteFile(path, logger)
			}

			// Do the "list" action if nothing else was set.
			return printFilePath(path, outWriter)
		})
}
