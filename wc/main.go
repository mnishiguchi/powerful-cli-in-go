package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

/*
Counts the number of words in a given text.

## Examples

    # Count words
    ❯ cat main.go | ./wc
    125

    # Count lines
    ❯ cat main.go | ./wc -l
    48
*/
func main() {
	// Define a boolean flag -l to count lines instead of words.
	forLines := flag.Bool("l", false, "Count lines")

	// Parse the flags provided by the user.
	flag.Parse()

	// Print the word count.
	fmt.Println(count(os.Stdin, *forLines))
}

func count(r io.Reader, forLines bool) int {
	// Prepare a scanner that reads text from a reader (such as files).
	scanner := bufio.NewScanner(r)

	// Determine the scanning behavior. See https://pkg.go.dev/bufio
	if forLines {
		scanner.Split(bufio.ScanLines)
	} else {
		scanner.Split(bufio.ScanWords)
	}

	// Count all the scanned words.
	wc := 0
	for scanner.Scan() {
		wc++
	}

	return wc
}
