package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
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

func deleteFile(path string, logger *log.Logger) error {
	if err := os.Remove(path); err != nil {
		return err
	}

	logger.Println(path)
	return nil
}

// Finds or creates the destination directory and creates the compressed archive.
func archiveFile(destDir, searchRoot, fileToArchive string) error {
	// Get info on the destination directory
	info, err := os.Stat(destDir)
	if err != nil {
		return err
	}

	fmt.Println(searchRoot, fileToArchive)

	// Check if the destination directory is actually a directory.
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", destDir)
	}

	// Determine the relative directory of the file to be archived in relation to
	// its source root path. This relative directory is need to create a similar
	// directory tree in the destination directory.
	relDir, err := filepath.Rel(searchRoot, filepath.Dir(fileToArchive))
	if err != nil {
		return err
	}

	// Create the target directory. By using the functions from the filepath
	// package, we ensure that the paths are build in accordance with the OS where
	// program is running.
	destName := fmt.Sprintf("%s.gz", filepath.Base(fileToArchive))
	targetPath := filepath.Join(destDir, relDir, destName)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	//
	destFile, err := os.OpenFile(targetPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(fileToArchive)
	if err != nil {
		return nil
	}
	defer srcFile.Close()

	// Use the gzip.Writer to create the compressed archive.
	gzWriter := gzip.NewWriter(destFile)
	// Store the source file name in the compressed archive.
	gzWriter.Name = filepath.Base(fileToArchive)
	// Create the compressed archive for the source file.
	if _, err = io.Copy(gzWriter, srcFile); err != nil {
		return err
	}
	// We do not defer this because we want to ensure we return any potential
	// errors. If the comressing fails, the calling function will get an error and
	// decide how to proceed.
	if err := gzWriter.Close(); err != nil {
		return err
	}

	return destFile.Close()
}
