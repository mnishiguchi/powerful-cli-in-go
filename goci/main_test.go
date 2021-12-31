package main

import (
	"bytes"
	"errors"
	"testing"
)

func TestRun(t *testing.T) {
	var testCases = []struct {
		name        string
		projectDir  string
		expected    string
		expectedErr error
	}{
		{name: "success",
			projectDir:  "./testdata/tool/",
			expected:    "go build: SUCCESS\n",
			expectedErr: nil},
		{name: "fail",
			projectDir:  "./testdata/toolErr/",
			expected:    "",
			expectedErr: &stepErr{step: "go build"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// A buffer to capture the output
			var outWriter bytes.Buffer

			err := run(tc.projectDir, &outWriter)

			// When an error is expected
			if tc.expectedErr != nil {
				if err == nil {
					t.Errorf("Expected error: %q. Got 'nil' instead.", tc.expectedErr)
					return
				}

				if !errors.Is(err, tc.expectedErr) {
					t.Errorf("Expected error: %q. Got %q", tc.expectedErr, err)
				}

				return
			}

			// When no error is expected
			if err != nil {
				t.Errorf("Unexpected error: %q", err)
			}

			if outWriter.String() != tc.expected {
				t.Errorf("Expected output: %q. Got %q", tc.expected, outWriter.String())
			}
		})
	}
}
