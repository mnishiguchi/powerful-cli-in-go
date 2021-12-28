package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"

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

/*
## Examples

    # Parse markdown and open a preview
    ./mdp -file README.md

    # Skip auto-preview
    ./mdp -file README.md -s
*/
func main() {
	// Define and parse flags
	filename := flag.String("file", "", "Markdown file to preview")
	skipPreview := flag.Bool("s", false, "Skip auto-preview")
	flag.Parse()

	// Print the usage in case wrong flags are provided.
	if *filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Do the work.
	if err := doWork(*filename, os.Stdout, *skipPreview); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Coordinates the execution of the multiple operations:
//   * Receives a markdown file
//   * parses it into HTML
//   * save the HTML to a new file
func doWork(markdownFile string, outWriter io.Writer, skipPreview bool) error {
	markdownData, err := ioutil.ReadFile(markdownFile)
	if err != nil {
		return err
	}

	htmlData := parseMarkdown(markdownData)

	// Create a temporary file
	temp, err := ioutil.TempFile("", "mdp*.html")
	if err != nil {
		return err
	}

	// Close the temporary file because we are not writing any data to it
	if err := temp.Close(); err != nil {
		return err
	}

	// Print the temporary file name and save HTML to that file
	outfile := temp.Name()
	fmt.Fprintln(outWriter, outfile)

	if err := saveHTML(outfile, htmlData); err != nil {
		return err
	}

	if skipPreview {
		return nil
	}

	return preview(outfile)
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

// Previews the file in a browser.
func preview(filename string) error {
	cName := ""
	cParams := []string{}

	// Define executable based on OS
	switch runtime.GOOS {
	case "linux":
		cName = "xdg-open"
	case "windows":
		cName = "cmd.exe"
		cParams = []string{"/C", "start"}
	case "darwin":
		cName = "open"
	default:
		return fmt.Errorf("OS not supported")
	}

	// Append filename to parameters slice
	cParams = append(cParams, filename)

	// Locate  executable in PATH
	cPath, err := exec.LookPath(cName)
	if err != nil {
		return err
	}

	// Open the file using the OS default program
	return exec.Command(cPath, cParams...).Run()
}
