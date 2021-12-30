package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
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

// Benchmarks the run() function.
//
// ## Examples
//
//   # Execute the benchmark
//   go test -bench . -run ^$
//   go test -bench . -run ^$ -benchtime=10x
//   go test -bench . -run ^$ -benchtime=10x | tee log/bench_results_00.txt
//
//   # Execute the CPU profiler
//   go test -bench . -run ^$ -benchtime=10x -cpuprofile log/cpu00.pprof
//   go tool pprof cpu00.pprof
//
//   # Execute the memory profiler
//   go test -bench . -run ^$ -benchtime=10x -memprofile log/mem00.pprof
//   go tool pprof -alloc_space log/mem00.pprof
//
//   # Display the total memory allocation
//   go test -bench . -run ^$ -benchtime=10x -benchmem | tee log/bench_results_00m.txt
//
//   # Trace our program
//   go test -bench . -run ^$ -benchtime=10x -trace log/trace01.out
//   go tool trace log/trace01.out
//
func BenchmarkRun(b *testing.B) {
	fileNames, err := filepath.Glob("./testdata/benchmark/*.csv")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := run(fileNames, "avg", 2, io.Discard); err != nil {
			b.Error(err)
		}
	}
}
