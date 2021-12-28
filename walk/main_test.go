package main

import (
	"bytes"
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
			cfg:      config{ext: "", size: 0, list: true},
			expected: "testdata/dir.log\n" + "testdata/dir2/script.sh\n"},
		{name: "FilterExtensionMatch",
			rootDir:  "testdata",
			cfg:      config{ext: ".log", size: 0, list: true},
			expected: "testdata/dir.log\n"},
		{name: "FilterExtensionSizeMatch",
			rootDir:  "testdata",
			cfg:      config{ext: ".log", size: 10, list: true},
			expected: "testdata/dir.log\n"},
		{name: "FilterExtensionSizeNoMatch",
			rootDir:  "testdata",
			cfg:      config{ext: ".log", size: 20, list: true},
			expected: ""},
		{name: "FilterExtensionNoMatch",
			rootDir:  "testdata",
			cfg:      config{ext: ".gz", size: 0, list: true},
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
