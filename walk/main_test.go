package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

			filesLeft, err := os.ReadDir(tempDir)
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

func TestRunArchive(t *testing.T) {
	testCases := []struct {
		name         string
		cfg          config
		extNoArchive string
		nArchive     int
		nNoArchive   int
	}{
		{name: "ArchiveExtensionNoMatch",
			cfg:          config{ext: ".log"},
			extNoArchive: ".gz",
			nArchive:     0,
			nNoArchive:   10},
		{name: "ArchiveExtensionMatch",
			cfg:          config{ext: ".log"},
			extNoArchive: "",
			nArchive:     10,
			nNoArchive:   0},
		{name: "ArchiveExtensionMixed",
			cfg:          config{ext: ".log"},
			extNoArchive: ".gz",
			nArchive:     5,
			nNoArchive:   5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Buffer to capture the output
			var buffer bytes.Buffer

			// Create a temporary directory with some files for testing.
			tempDir, cleanupTempDir := createTempDir(t, map[string]int{
				tc.cfg.ext:      tc.nArchive,
				tc.extNoArchive: tc.nNoArchive,
			})
			defer cleanupTempDir()

			// Create a archive directory with no files.
			archiveDir, cleanupArchive := createTempDir(t, nil)
			defer cleanupArchive()

			// Pass the archive directory to the config struct before invoking the
			// run() function.
			tc.cfg.archiveDir = archiveDir

			// Invoke the run() function, which will output to the buffer.
			if err := run(tempDir, &buffer, tc.cfg); err != nil {
				t.Fatal(err)
			}

			// Validate the output assuming the run functions completes successfully.
			// Since the test create the directory and files dynamically for each test.
			// we do not have the name of the files beforehand.
			//
			// Find all the file names from the temporary directory that match the
			// archiving extension.
			fileNamePattern := filepath.Join(tempDir, fmt.Sprintf("*%s", tc.cfg.ext))

			expectedFiles, err := filepath.Glob(fileNamePattern)
			if err != nil {
				t.Fatal(err)
			}

			// Verify the output.
			expectedOutput := strings.Join(expectedFiles, "\n")
			actualOutput := strings.TrimSpace(buffer.String()) // remove the last new line
			if expectedOutput != actualOutput {
				t.Errorf("Expected %q, got %q instead\n", expectedOutput, actualOutput)
			}

			// Read the content of the temprary archive directory.
			filesArchived, err := os.ReadDir(archiveDir)
			if err != nil {
				t.Fatal(err)
			}

			// Verify the number of files archived.
			if len(filesArchived) != tc.nArchive {
				t.Errorf("Expected %d files archived, got %d instead\n", tc.nArchive, len(filesArchived))
			}
		})
	}
}

// A test helper that creates a temporary directory and create files in that
// directory based on the provided map.
func createTempDir(t *testing.T, extToHowMany map[string]int) (dirname string, cleanup func()) {
	t.Helper() // Mark this test as a test helper

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "walktest")
	if err != nil {
		t.Fatal(err)
	}

	for ext, howMany := range extToHowMany {
		for j := 1; j <= howMany; j++ {
			fileName := fmt.Sprintf("file%d%s", j, ext)
			filePath := filepath.Join(tempDir, fileName)
			if err := os.WriteFile(filePath, []byte("dummy"), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}

	return tempDir, func() { os.RemoveAll(tempDir) }
}
