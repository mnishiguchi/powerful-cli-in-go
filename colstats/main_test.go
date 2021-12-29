package main

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	testCases := []struct {
		name        string
		column      int
		op          string
		expected    string
		files       []string
		expectedErr error
	}{
		{name: "RunAvgOneFile",
			column:      3,
			op:          "avg",
			expected:    "227.6\n",
			files:       []string{"./testdata/example.csv"},
			expectedErr: nil},
		{name: "RunAvgMultipleFile",
			column:      3,
			op:          "avg",
			expected:    "233.84\n",
			files:       []string{"./testdata/example.csv", "./testdata/example2.csv"},
			expectedErr: nil},
		{name: "RunFailRead",
			column:      2,
			op:          "avg",
			expected:    "",
			files:       []string{"./testdata/example.csv", "./testdata/fakefile.csv"},
			expectedErr: os.ErrNotExist},
		{name: "RunFailColumn",
			column:      0,
			op:          "avg",
			expected:    "",
			files:       []string{"./testdata/example.csv"},
			expectedErr: ErrInvalidColumn},
		{name: "RunFailNoColumn",
			column:      0,
			op:          "avg",
			expected:    "",
			files:       []string{},
			expectedErr: ErrNoFile},
		{name: "RunFailOperation",
			column:      2,
			op:          "invalid",
			expected:    "",
			files:       []string{"./testdata/example.csv"},
			expectedErr: ErrInvalidOperation},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture the output
			var outputWriter bytes.Buffer

			err := run(tc.files, tc.op, tc.column, &outputWriter)

			if tc.expectedErr != nil {
				if err == nil {
					t.Errorf("Expected error. Got nil instead")
				}

				if !errors.Is(err, tc.expectedErr) {
					t.Errorf("Expected error %q, got %q instead", tc.expectedErr, err)
				}

				return
			}

			if err != nil {
				t.Errorf("Unexpected err: %q", err)
			}

			if outputWriter.String() != tc.expected {
				t.Errorf("Expected %q, got %q instead", tc.expected, &outputWriter)
			}
		})
	}
}
