package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

const (
	header = `<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="content-type" content="text/html; charset=utf-8">
    <title>Markdown Preview</title>
  </head>
  <body>
`
	footer = `
  </body>
</html>
`
)

func main() {
	// Define and parse flags
	filename := flag.String("file", "", "Markdown file to preview")
	flag.Parse()

	// Print the usage in case wrong flags are provided.
	if *filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Do the work.
	if err := doWork(*filename); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Coordinates the execution of the multiple operations:
//   * Receives a markdown file
//   * parses it into HTML
//   * save the HTML to a new file
func doWork(markdownFile string) error {
	markdownData, err := ioutil.ReadFile(markdownFile)

	if err != nil {
		return err
	}

	htmlData := parseMarkdown(markdownData)

	// As an example, for "example.md" the outfile would be "example.md.html"
	outfile := fmt.Sprintf("%s.html", filepath.Base(markdownFile))

	return saveHTML(outfile, htmlData)
}

// Converts Markdown data to HTML data.
func parseMarkdown(markdownData []byte) []byte {
	// https://github.com/russross/blackfriday
	html := blackfriday.Run(markdownData)
	// https://github.com/microcosm-cc/bluemonday
	sanitizedHTML := bluemonday.UGCPolicy().SanitizeBytes(html)

	// Create a buffer of bytes to write to a file
	var buffer bytes.Buffer

	buffer.WriteString(header)  // string
	buffer.Write(sanitizedHTML) // bytes
	buffer.WriteString(footer)  // string

	return buffer.Bytes()
}

// Saves HTML data to a specified outfile.
func saveHTML(outfile string, htmlData []byte) error {
	// The 644 file permission is for creating a file that is both reacable and
	// writable by the owner but only readable by anyone else.
	return ioutil.WriteFile(outfile, htmlData, 0644)
}
