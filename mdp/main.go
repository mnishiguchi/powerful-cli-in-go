package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

const (
	defaultTemplate = `<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="content-type" content="text/html; charset=utf-8">
    <title>{{ .Title }}</title>
  </head>
  <body>
{{ .Body }}
  </body>
</html>
`
)

// Represents the HTML content to add into the template.
type templateProps struct {
	Title string
	// Never use HTML from untrusted sources as it could present a security risk.
	Body template.HTML
}

/*
## Examples

    # Parse markdown using default HTML template and open a preview
    ./mdp -file README.md

    # Parse markdown using custom HTML template and open a preview
    ./mdp -file README.md -t template-fmt.html.tmpl

    # Skip auto-preview
    ./mdp -file README.md -s
*/
func main() {
	// Define and parse flags
	filename := flag.String("file", "", "Markdown file to preview")
	skipPreview := flag.Bool("s", false, "Skip auto-preview")
	templateFile := flag.String("t", "", "Alternative template name")
	flag.Parse()

	// Print the usage in case wrong flags are provided.
	if *filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Do the work.
	if err := doWork(*filename, *templateFile, os.Stdout, *skipPreview); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Coordinates the execution of the multiple operations:
//   * Receives a markdown file
//   * parses it into HTML
//   * save the HTML to a new file
func doWork(markdownFile string, templateFile string, outWriter io.Writer, skipPreview bool) error {
	markdownData, err := os.ReadFile(markdownFile)
	if err != nil {
		return err
	}

	htmlData, err := parseMarkdown(markdownData, templateFile)
	if err != nil {
		return err
	}

	// Create a temporary file
	temp, err := os.CreateTemp("", "mdp*.html")
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

	// It is our responsibility to delete the temporary files.
	// Ensure that the outfile is deleted when the current function returns.
	defer os.Remove(outfile)

	return preview(outfile)
}

// Converts Markdown data to HTML data.
func parseMarkdown(markdownData []byte, templateFile string) ([]byte, error) {
	// https://github.com/russross/blackfriday
	html := blackfriday.Run(markdownData)
	// https://github.com/microcosm-cc/bluemonday
	sanitizedHTML := bluemonday.UGCPolicy().SanitizeBytes(html)

	// Parse the default template into a new template.
	// By using this approach, we always have the default template ready to execute.
	parsedTemplate, err := template.New("mdp").Parse(defaultTemplate)
	if err != nil {
		return nil, err
	}

	// If the user provides a template file, use that template instead.
	if templateFile != "" {
		parsedTemplate, err = template.ParseFiles(templateFile)
		if err != nil {
			return nil, err
		}
	}

	// Create a buffer of bytes to write to a file
	var buffer bytes.Buffer

	// Replace the placeholders and write the result to the buffer
	err = parsedTemplate.Execute(&buffer, templateProps{
		Title: "Markdown Preview",
		Body:  template.HTML(sanitizedHTML),
	})
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// Saves HTML data to a specified outfile.
func saveHTML(outfile string, htmlData []byte) error {
	// The 644 file permission is for creating a file that is both reacable and
	// writable by the owner but only readable by anyone else.
	return os.WriteFile(outfile, htmlData, 0644)
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
	err = exec.Command(cPath, cParams...).Run()

	// Give the browser some time to open the file before deleting it.
	// Adding a delay is not a recommended long-term solution. This is a quick fix
	// so we can focus on the cleanup functionality.
	time.Sleep(2 * time.Second)

	return err
}
