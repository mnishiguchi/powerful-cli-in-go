package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"
	"testing/iotest"
)

// Tests all the operation functions.
func TestOperations(t *testing.T) {
	data := [][]float64{
		{10, 20, 15, 30, 45, 50, 100, 30},
		{5.5, 8, 2.2, 9.75, 8.45, 3, 2.5, 10.25, 4.75, 6.1, 7.67, 12.287, 5.47},
		{-10, -20},
		{102, 37, 44, 57, 67, 129},
	}

	testCases := []struct {
		name     string
		opFunc   statsFunc
		expected []float64
	}{
		{"Sum", sum, []float64{300, 85.927, -30, 436}},
		{"Avg", avg, []float64{37.5, 6.609769230769231, -15, 72.66666666666667}},
	}

	for _, tc := range testCases {
		for index, expected := range tc.expected {
			name := fmt.Sprintf("%sData%d", tc.name, index)
			t.Run(name, func(t *testing.T) {
				actual := tc.opFunc(data[index])

				if !veryclose(actual, expected) {
					t.Errorf("Expected %g, got %g instead", expected, actual)
				}
			})
		}
	}
}

func TestCSV2Float(t *testing.T) {
	csvData := `IP Address,Requests,Response Time
192.168.0.199,2056,236
192.168.0.88,899,220
192.168.0.199,3054,226
192.168.0.100,4133,218
192.168.0.199,950,238
`
	testCases := []struct {
		name        string
		column      int
		expected    []float64
		expectedErr error
		inputReader io.Reader
	}{
		{name: "Column2", column: 2,
			expected:    []float64{2056, 899, 3054, 4133, 950},
			expectedErr: nil,
			inputReader: bytes.NewBufferString(csvData),
		},
		{name: "Column3", column: 3,
			expected:    []float64{236, 220, 226, 218, 238},
			expectedErr: nil,
			inputReader: bytes.NewBufferString(csvData),
		},
		// Simulate a reading failure.
		{name: "FailRead", column: 1,
			expected:    nil,
			expectedErr: iotest.ErrTimeout,
			inputReader: iotest.TimeoutReader(bytes.NewReader([]byte{0})),
		},
		{name: "FailNotNumber", column: 1,
			expected:    nil,
			expectedErr: ErrNotNumber,
			inputReader: bytes.NewBufferString(csvData),
		},
		{name: "FailInvalidColumn", column: 4,
			expected:    nil,
			expectedErr: ErrInvalidColumn,
			inputReader: bytes.NewBufferString(csvData),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := csv2float(tc.inputReader, tc.column)

			if tc.expectedErr != nil {
				if err == nil {
					t.Errorf("Expected error. Got nil instead")
				}

				if !errors.Is(err, tc.expectedErr) {
					t.Errorf("Expected error %q, got %q instead", tc.expectedErr, err)
				}

				// Prevent the execution of the remaining checks.
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %q", err)
			}

			for i, expected := range tc.expected {
				if actual[i] != expected {
					t.Errorf("Expected %g, got %g instead", expected, actual[i])
				}
			}
		})
	}
}

// Compares floating-point numbers.
// See https://go.dev/src/math/all_test.go
func tolerance(a, b, e float64) bool {
	// Multiplying by e here can underflow denormal values to zero.
	// Check a==b so that at least if a and b are small and identical
	// we say they match.
	if a == b {
		return true
	}
	d := a - b
	if d < 0 {
		d = -d
	}

	// note: b is correct (expected) value, a is actual value.
	// make error tolerance a fraction of b, not a.
	if b != 0 {
		e = e * b
		if e < 0 {
			e = -e
		}
	}
	return d < e
}
func close(a, b float64) bool      { return tolerance(a, b, 1e-14) }
func veryclose(a, b float64) bool  { return tolerance(a, b, 4e-16) }
func soclose(a, b, e float64) bool { return tolerance(a, b, e) }
