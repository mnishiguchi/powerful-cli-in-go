package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

// A generic statistical function
type statsFunc func(data []float64) float64

func sum(data []float64) float64 {
	sum := 0.0

	for _, v := range data {
		sum += v
	}

	return sum
}

func avg(data []float64) float64 {
	return sum(data) / float64(len(data))
}

// Convert values at the specified column into float64.
func csv2float(inputReader io.Reader, column int) ([]float64, error) {
	// Create a CSV reader that reads data from CSV files
	csvReader := csv.NewReader(inputReader)

	// Convert one-based index to zero-based index
	column--

	allRecords, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("Cannot read data from file: %w", err)
	}

	var data []float64

	for i, row := range allRecords {
		// Ignore the title row.
		if i == 0 {
			continue
		}

		// The column number out of bounds
		if len(row) <= column {
			return nil, fmt.Errorf("%w: File has only %d columns", ErrInvalidColumn, len(row))
		}

		// Convert a string value to float64
		value, err := strconv.ParseFloat(row[column], 64)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrNotNumber, err)
		}

		data = append(data, value)
	}

	return data, nil
}
