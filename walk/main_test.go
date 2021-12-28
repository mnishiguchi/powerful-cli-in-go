package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	testCases := []struct {
		name     string
		rootDir  string
		cfg      config
		expected string
	}{
		{name: "NoFilter",
			rootDir:  "testdata",
			cfg:      config{ext: "", size: 0, isListAction: true},
			expected: "testdata/dir.log\n" + "testdata/dir2/script.sh\n"},
		{name: "FilterExtensionMatch",
			rootDir:  "testdata",
			cfg:      config{ext: ".log", size: 0, isListAction: true},
			expected: "testdata/dir.log\n"},
		{name: "FilterExtensionSizeMatch",
			rootDir:  "testdata",
			cfg:      config{ext: ".log", size: 10, isListAction: true},
			expected: "testdata/dir.log\n"},
		{name: "FilterExtensionSizeNoMatch",
			rootDir:  "testdata",
			cfg:      config{ext: ".log", size: 20, isListAction: true},
			expected: ""},
		{name: "FilterExtensionNoMatch",
			rootDir:  "testdata",
			cfg:      config{ext: ".gz", size: 0, isListAction: true},
			expected: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buffer bytes.Buffer
			if err := run(tc.rootDir, &buffer, tc.cfg); err != nil {
				t.Fatal(err)
			}

			res := buffer.String()

			if tc.expected != res {
				t.Errorf("Expected %q, got %q instead\n", tc.expected, res)
			}
		})
	}
}

func TestRunDeleteFiles(t *testing.T) {
	testCases := []struct {
		name        string
		cfg         config
		extNoDelete string
		nDelete     int
		nNoDelete   int
		expected    string
	}{
		{name: "DeleteFilesNoMatch",
			cfg:         config{ext: ".log", isDeleteAction: true},
			extNoDelete: "",
			nDelete:     0,
			nNoDelete:   10,
			expected:    ""},
		{name: "DeleteFilesMatch",
			cfg:         config{ext: ".log", isDeleteAction: true},
			extNoDelete: ".gz",
			nDelete:     10,
			nNoDelete:   0,
			expected:    ""},
		{name: "DeleteFilesMixed",
			cfg:         config{ext: ".log", isDeleteAction: true},
			extNoDelete: ".gz",
			nDelete:     5,
			nNoDelete:   5,
			expected:    ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				buffer    bytes.Buffer
				logBuffer bytes.Buffer
			)

			// Assign the address of the log buffer to the log writer field of the config.
			tc.cfg.logWriter = &logBuffer

			tempDir, cleanupTempDir := createTempDir(t, map[string]int{
				tc.cfg.ext:     tc.nDelete,
				tc.extNoDelete: tc.nNoDelete,
			})
			defer cleanupTempDir()

			if err := run(tempDir, &buffer, tc.cfg); err != nil {
				t.Fatal(err)
			}

			res := buffer.String()

			if tc.expected != res {
				t.Errorf("Expected %q, got %q instead\n", tc.expected, res)
			}

			filesLeft, err := ioutil.ReadDir(tempDir)
			if err != nil {
				t.Errorf("Expected %d files left, got %d instead\n", tc.nNoDelete, len(filesLeft))
			}

			expLogLines := tc.nDelete + 1
			lines := bytes.Split(logBuffer.Bytes(), []byte("\n"))
			if len(lines) != expLogLines {
				t.Errorf("Expected %d log lines, got %d instead\n", expLogLines, len(lines))
			}
		})
	}

}

// A test helper that creates a temporary directory.
func createTempDir(t *testing.T, extToHowMany map[string]int) (dirname string, cleanup func()) {
	t.Helper() // Mark this test as a test helper

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "walktest")
	if err != nil {
		t.Fatal(err)
	}

	for ext, howMany := range extToHowMany {
		for j := 1; j <= howMany; j++ {
			fileName := fmt.Sprintf("file%d%s", j, ext)
			filePath := filepath.Join(tempDir, fileName)
			if err := ioutil.WriteFile(filePath, []byte("dummy"), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}

	return tempDir, func() { os.RemoveAll(tempDir) }
}
