package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

// ## Examples
//
//    go run . -col 1 tmp/example.csv
//
func main() {
	// Verify and parse arguments
	op := flag.String("op", "sum", "Operation to be executed")
	col := flag.Int("col", 1, "CSV column on which to execute the operation")
	flag.Parse()

	if err := run(flag.Args(), *op, *col, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Coordinates the entire program's execution.
func run(fileNames []string, op string, column int, out io.Writer) error {
	var operationFunc statsFunc

	// No file to process.
	if len(fileNames) == 0 {
		return ErrNoFile
	}

	// The column number must be greater than one.
	if column < 1 {
		return fmt.Errorf("%w: %d", ErrInvalidColumn, column)
	}

	// Determine the operation.
	switch op {
	case "sum":
		operationFunc = sum
	case "avg":
		operationFunc = avg
	default:
		return fmt.Errorf("%w: %s", ErrInvalidOperation, op)
	}

	// A slice we use to consolidated the parsed values from all the files.
	consolidatedData := make([]float64, 0)

	for _, fileName := range fileNames {
		f, err := os.Open(fileName)
		if err != nil {
			return fmt.Errorf("Cannot open file: %w", err)
		}

		// Parse the CSV into a slice of float64 to consolidate.
		values, err := csv2float(f, column)
		if err != nil {
			return err
		}

		if err := f.Close(); err != nil {
			return err
		}

		consolidatedData = append(consolidatedData, values...)
	}

	// Execute the user-specified operation and print out the results.
	_, err := fmt.Fprintln(out, operationFunc(consolidatedData))

	// return any potential error.
	return err
}
