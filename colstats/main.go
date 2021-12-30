package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
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

	// Create the channel to receive results or errors of operations.
	// Using an empty struct is a common pattern when we do not need to send any data.
	chFileName := make(chan string)   // a queue for file names to be processed
	chResults := make(chan []float64) // results of processing each file
	chErr := make(chan error)         // potential errors
	chDone := make(chan struct{})     // notify when all files have been processed

	// The WaitGroup provides a mechanism to coordinate the goroutine execution.
	csvWorkerWaitGroup := sync.WaitGroup{}

	go func() {
		// At the end, close the channel indicating no more work is left to do.
		defer close(chFileName)

		// Push all the file names to the chFileName channel so each one will be
		// processed when a worker is available.
		for _, fileName := range fileNames {
			chFileName <- fileName
		}
	}()

	// Create a worker goroutine per CPU.
	for i := 0; i < runtime.NumCPU(); i++ {
		// Increment the WaitGroup counter to indicate a running goroutine.
		csvWorkerWaitGroup.Add(1)

		// A worker goroutine.
		go func() {
			// Decrement the WaitGroup counter when the function finishes.
			defer csvWorkerWaitGroup.Done()

			// This loop gets values from the chFileName channel until it is closed.
			for fileName := range chFileName {
				// Open the file for reading.
				f, err := os.Open(fileName)
				if err != nil {
					chErr <- fmt.Errorf("Cannot open file: %w", err)
					return
				}

				// Parse the CSV into a slice of float64 to consolidate.
				results, err := csv2float(f, column)
				if err != nil {
					chErr <- err
				}

				if err := f.Close(); err != nil {
					chErr <- err
				}

				chResults <- results
			}
		}()
	}

	go func() {
		// Wait until all files have been processed.
		csvWorkerWaitGroup.Wait()

		// Close the chDone channel signaling that the process is complete.
		close(chDone)
	}()

	// This blocks the execution of the program until any of the channels is
	// ready to communicate.
	for {
		// The conditions are communication operations through a channel.
		select {
		case err := <-chErr:
			return err
		case results := <-chResults:
			// In Go, it is idiomatic to use a channel to communicate the values
			// between goroutines so we can avoid a date race condition.
			consolidatedData = append(consolidatedData, results...)
		case <-chDone:
			// Execute the user-specified operation and print out the results.
			_, err := fmt.Fprintln(out, operationFunc(consolidatedData))
			// return any potential error.
			return err
		}
	}
}
