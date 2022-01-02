package scan_test

import (
	"net"
	"strconv"
	"testing"

	"mnishiguchi.com/pscan/scan"
)

func TestStateString(t *testing.T) {
	ps := scan.ScannedPort{}

	if ps.Open.String() != "closed" {
		t.Errorf("Expected %q, got %q instead\n", "closed", ps.Open.String())
	}

	ps.Open = true
	if ps.Open.String() != "open" {
		t.Errorf("Expected %q, got %q instead\n", "open", ps.Open.String())
	}
}

func TestRunHostFound(t *testing.T) {
	testCases := []struct {
		name          string
		expectedState string
	}{
		{"open port", "open"},
		{"closed port", "closed"},
	}

	hostName := "localhost"

	// Initialize host list
	hl := &scan.HostList{}
	hl.Add(hostName)

	// Initialize ports (1 open, 1 closed)
	ports := []int{}
	for _, tc := range testCases {
		// Use the port number 0 so an available port on the host is automatically chosen.
		ln, err := net.Listen("tcp", net.JoinHostPort(hostName, "0"))
		if err != nil {
			t.Fatal(err)
		}
		defer ln.Close()
		// fmt.Println(ln.Addr()) // debug

		// Extract the port number string from the listener address.
		_, portStr, err := net.SplitHostPort(ln.Addr().String())
		if err != nil {
			t.Fatal(err)
		}

		// Convert portStr into integer.
		port, err := strconv.Atoi(portStr)
		if err != nil {
			t.Fatal(err)
		}

		// Push the port numer to the port list.
		ports = append(ports, port)
		if tc.name == "closed port" {
			ln.Close()
		}
	}

	// Execute the Run method.
	scanResults := scan.Run(hl, ports)

	if len(scanResults) != 1 {
		t.Fatalf("Expected 1 result, got %d instead\n", len(scanResults))
	}

	if scanResults[0].Host != hostName {
		t.Errorf("Expected host %q, got %q instead\n", hostName, scanResults[0].Host)
	}

	if scanResults[0].NotFound {
		t.Errorf("Expected host %q to be found\n", hostName)
	}

	if len(scanResults[0].ScannedPorts) != 2 {
		t.Fatalf("Expected 2 port states, got %d instead\n", len(scanResults[0].ScannedPorts))
	}

	// Loop through test cases verifying each port state.
	for i, tc := range testCases {
		if scanResults[0].ScannedPorts[i].Port != ports[i] {
			t.Errorf("Expected port %d, got %d instead\n", ports[0], scanResults[0].ScannedPorts[i].Port)
		}

		if scanResults[0].ScannedPorts[i].Open.String() != tc.expectedState {
			t.Errorf("Expected port %d to be %s\n", ports[i], tc.expectedState)
		}
	}
}

func TestRunHostNotFound(t *testing.T) {
	// Name resolution on this host should fail unless we have it on our DNS.
	host := "389.389.389.389"
	hl := &scan.HostList{}
	hl.Add(host)

	scanResults := scan.Run(hl, []int{})

	if len(scanResults) != 1 {
		t.Fatalf("Expected 1 result, got %d instead\n", len(scanResults))
	}

	if scanResults[0].Host != host {
		t.Errorf("Expected host %q, got %q instead\n", host, scanResults[0].Host)
	}

	if !scanResults[0].NotFound {
		t.Errorf("Expected host %q not to be found\n", host)
	}

	if len(scanResults[0].ScannedPorts) != 0 {
		t.Fatalf("Expected 0 port states, got %d instead\n", len(scanResults[0].ScannedPorts))
	}
}
