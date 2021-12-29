package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

const (
	inputFile  = "./testdata/test1.md"
	goldenFile = "./testdata/test1.md.html"
)

func TestParseContent(t *testing.T) {
	markdownData, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatal(err)
	}

	result, err := parseMarkdown(markdownData, "")
	if err != nil {
		t.Fatal(err)
	}

	expected, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(expected, result) {
		t.Logf("golden:\n%s\n", expected)
		t.Logf("result:\n%s\n", result)
		t.Error("Result content does not match golden file")
	}
}

func TestDoWork(t *testing.T) {
	// Captures the outfile name that the doWork() function prints.
	var mockStdout bytes.Buffer

	// Do all the work for converting Markdown to HTML and save it to a file.
	skipPreview := true
	if err := doWork(inputFile, "", &mockStdout, skipPreview); err != nil {
		t.Fatal(err)
	}

	// Read the resulting HTML from the outfile. The output contains a new line
	// at the end so we want to remove it.
	outfile := strings.TrimSpace(mockStdout.String())
	result, err := os.ReadFile(outfile)
	if err != nil {
		t.Fatal(err)
	}

	// Read the golden file.
	expected, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(expected, result) {
		t.Logf("golden:\n%s\n", expected)
		t.Logf("result:\n%s\n", result)
		t.Error("Result content does not match golden file")
	}

	os.Remove(outfile)
}
