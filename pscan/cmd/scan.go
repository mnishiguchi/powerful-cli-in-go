/*
Copyright © 2022 Masatoshi Nishiguchi

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"mnishiguchi.com/pscan/scan"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Run a port scan on the hosts",
	RunE: func(cmd *cobra.Command, args []string) error {
		hostsFile, err := cmd.Flags().GetString("hosts-file")
		if err != nil {
			return err
		}

		ports, err := cmd.Flags().GetIntSlice("ports")
		if err != nil {
			return err
		}

		return scanAction(os.Stdout, hostsFile, ports)
	},
}

func scanAction(outWriter io.Writer, hostsFile string, ports []int) error {
	hl := &scan.HostList{}

	if err := hl.Load(hostsFile); err != nil {
		return nil
	}

	results := scan.Run(hl, ports)

	return printResults(outWriter, results)
}

func printResults(outWriter io.Writer, results []scan.Results) error {
	var message strings.Builder

	for _, r := range results {
		message.WriteString(fmt.Sprintf("%s:", r.Host))
		if r.NotFound {
			message.WriteString(fmt.Sprintf(" Host not found\n\n"))
			continue
		}
		message.WriteString("\n")

		for _, p := range r.ScannedPorts {
			message.WriteString(fmt.Sprintf("\t%d: %s\n", p.Port, p.Open))
		}
		message.WriteString("\n")
	}

	_, err := fmt.Fprint(outWriter, message.String())
	return err
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scanCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	scanCmd.Flags().IntSliceP("ports", "p", []int{22, 80, 443}, "ports to scan")
}
