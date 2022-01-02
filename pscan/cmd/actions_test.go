package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"

	"mnishiguchi.com/pscan/scan"
)

/*
The setup sets up the test environment, creating a temporary file and initializing
a host list if required. It accepts an instance of the type testing.T, host
names that we initialize a list with, and a bool that indicaates whether the list
should be initialized. It returns the name of the temporary file and a cleanup
function that deletes the temporary file after it was used.
*/
func setup(t *testing.T, hostNames []string, shouldInitList bool) (string, func()) {
	// Create a temp file
	temp, err := os.CreateTemp("", "pscan")
	if err != nil {
		t.Fatal(err)
	}
	temp.Close()

	// Initialize list if needed
	if shouldInitList {
		hl := &scan.HostList{}

		// Add to the list each of the provided host names.
		for _, h := range hostNames {
			hl.Add(h)
		}

		if err := hl.Save(temp.Name()); err != nil {
			t.Fatal(err)
		}
	}

	// Return the temp file name and the cleanup function.
	return temp.Name(), func() {
		os.Remove(temp.Name())
	}
}

func TestHostActions(t *testing.T) {
	hostNames := []string{
		"host1",
		"host2",
		"host3",
	}

	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
		shouldInitList bool
		actionFn       func(io.Writer, string, []string) error
	}{
		{
			name:           "add action",
			args:           hostNames,
			expectedOutput: "Added host: host1\n" + "Added host: host2\n" + "Added host: host3\n",
			shouldInitList: false,
			actionFn:       addAction,
		},
		{
			name:           "list action",
			args:           hostNames,
			expectedOutput: "host1\n" + "host2\n" + "host3\n",
			shouldInitList: true,
			actionFn:       listAction,
		},
		{
			name:           "delete action",
			args:           []string{"host1", "host2"},
			expectedOutput: "Deleted host: host1\n" + "Deleted host: host2\n",
			shouldInitList: true,
			actionFn:       deleteAction,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			temp, cleanup := setup(t, hostNames, tc.shouldInitList)
			defer cleanup()

			// Define var to capture the action output
			var actualOutput bytes.Buffer

			// Execute the action and capture the output
			if err := tc.actionFn(&actualOutput, temp, tc.args); err != nil {
				t.Fatalf("Expected no error, got %q\n", err)
			}

			if actualOutput.String() != tc.expectedOutput {
				t.Errorf("Expected output %q, got %q\n", tc.expectedOutput, actualOutput.String())
			}
		})
	}
}

func TestScanAction(t *testing.T) {
	hosts := []string{
		"localhost",
		"unknownhostoutthere", // This does not exist in our network
	}

	// Set up the test with the host names.
	temp, cleanup := setup(t, hosts, true)
	defer cleanup()

	// Initialize ports (1 open, 1 closed)
	ports := []int{}

	for i := 0; i < 2; i++ {
		ln, err := net.Listen("tcp", net.JoinHostPort("localhost", "0"))
		if err != nil {
			t.Fatal(err)
		}
		defer ln.Close()

		_, portStr, err := net.SplitHostPort(ln.Addr().String())
		if err != nil {
			t.Fatal(err)
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			t.Fatal(err)
		}

		ports = append(ports, port)

		if i == 1 {
			ln.Close()
		}
	}

	expectedOut := fmt.Sprint("localhost:\n")
	expectedOut += fmt.Sprintf("\t%d: open\n", ports[0])
	expectedOut += fmt.Sprintf("\t%d: closed\n", ports[1])
	expectedOut += "\n"
	expectedOut += fmt.Sprint("unknownhostoutthere: Host not found\n")
	expectedOut += "\n"

	var outWriter bytes.Buffer

	if err := scanAction(&outWriter, temp, ports); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	if outWriter.String() != expectedOut {
		t.Errorf("Expected output %q, got %q\n", expectedOut, outWriter.String())
	}
}

// Executes all the operations in the defined sequence
// add -> list -> delete -> list -> scan
func TestIntegration(t *testing.T) {
	hostNames := []string{
		"host1",
		"host2",
		"host3",
	}

	temp, cleanup := setup(t, hostNames, false)
	defer cleanup()

	hostToDelete := "host2"

	hostsAfterDeletion := []string{
		"host1",
		"host3",
	}

	// Define buffers to capture the output
	var actualOutput, expectedOutput bytes.Buffer

	// add
	for _, v := range hostNames {
		expectedOutput.WriteString(fmt.Sprintf("Added host: %s\n", v))
	}
	// list
	expectedOutput.WriteString(strings.Join(hostNames, "\n"))
	expectedOutput.WriteString("\n")
	// delete
	expectedOutput.WriteString(fmt.Sprintf("Deleted host: %s\n", hostToDelete))
	// list
	expectedOutput.WriteString(strings.Join(hostsAfterDeletion, "\n"))
	expectedOutput.WriteString("\n")
	// scan
	for _, v := range hostsAfterDeletion {
		expectedOutput.WriteString(fmt.Sprintf("%s: Host not found\n", v))
		expectedOutput.WriteString("\n")
	}

	// add
	if err := addAction(&actualOutput, temp, hostNames); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	// list
	if err := listAction(&actualOutput, temp, hostNames); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	// delete
	if err := deleteAction(&actualOutput, temp, []string{hostToDelete}); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	// list
	if err := listAction(&actualOutput, temp, hostNames); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	// scan
	if err := scanAction(&actualOutput, temp, nil); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	// Verify the command output
	if actualOutput.String() != expectedOutput.String() {
		t.Errorf("expected output %q, got %q\n", expectedOutput.String(), actualOutput.String())
	}
}
