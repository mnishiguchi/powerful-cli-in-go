package scan

import (
	"fmt"
	"net"
	"time"
)

// Results reporesents the scan results for a single host.
type Results struct {
	Host         string
	NotFound     bool          // Whether the host can be resolved to a valid IP address
	ScannedPorts []ScannedPort // Each port scanned
}

// ScannedPort represents the state of a single TCP port.
type ScannedPort struct {
	Port int
	Open isOpen
}

// state indicates whether a port is open or closed.
type isOpen bool

// String converts the boolean value of state to a human readable string.
func (isOpen isOpen) String() string {
	if isOpen {
		return "open"
	}

	return "closed"
}

func Run(hl *HostList, ports []int) []Results {
	res := make([]Results, 0, len(hl.Hosts))

	for _, h := range hl.Hosts {
		r := Results{
			Host: h,
		}

		if _, err := net.LookupHost(h); err != nil {
			r.NotFound = true
			res = append(res, r)
			continue
		}

		for _, p := range ports {
			r.ScannedPorts = append(r.ScannedPorts, scanPort(h, p))
		}
		res = append(res, r)
	}

	return res
}

// scanPort performs a port scan on a single TCP port.
func scanPort(host string, port int) ScannedPort {
	p := ScannedPort{Port: port, Open: false}

	// Get the network address.
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	// Attempt to connect to a network address within one second.
	scanConn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err != nil {
		return p
	}
	scanConn.Close()

	// Now the port is open.
	p.Open = true

	return p
}
