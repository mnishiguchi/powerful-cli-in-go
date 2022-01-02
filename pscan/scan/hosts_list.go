package scan

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"sort"
)

// We will use these errors during tests.
var (
	ErrAlreadyExists = errors.New("Host already in the list")
	ErrNotExists     = errors.New("Host not in the list")
)

// HostList represents a list of hosts to run the port scan.
type HostList struct {
	Hosts []string
}

// Add adds a host from the list.
func (hl *HostList) Add(hostName string) error {
	if found, _ := hl.search(hostName); found {
		return fmt.Errorf("%w: %s", ErrAlreadyExists, hostName)
	}

	hl.Hosts = append(hl.Hosts, hostName)

	return nil
}

// Remove removes a host from the list.
func (hl *HostList) Remove(hostName string) error {
	if found, index := hl.search(hostName); found {
		hl.Hosts = append(hl.Hosts[:index], hl.Hosts[index+1:]...)

		return nil
	}

	return fmt.Errorf("%w: %s", ErrNotExists, hostName)
}

// Load obtains host names from a hosts file.
func (hl *HostList) Load(hostsFile string) error {
	// Open the file for reading.
	f, err := os.Open(hostsFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}
	defer f.Close()

	// Scanner that reads lines.
	lineScanner := bufio.NewScanner(f)

	// Push each line to the list.
	for lineScanner.Scan() {
		hl.Hosts = append(hl.Hosts, lineScanner.Text())
	}

	return nil
}

// Save saves hosts list to a hosts file.
func (hl *HostList) Save(hostsFile string) error {
	var output bytes.Buffer

	for _, h := range hl.Hosts {
		output.WriteString(fmt.Sprintln(h))
	}

	return os.WriteFile(hostsFile, output.Bytes(), 0644)
}

// Search searches searches for a host in the list.
func (hl *HostList) search(hostName string) (bool, int) {
	// Sort strings in ascending order because sort.SearchStrings() assumes that
	// the strings are sorted beforehand.
	sort.Strings(hl.Hosts)

	// Search and get the index.
	index := sort.SearchStrings(hl.Hosts, hostName)

	// Validate the index and determine the return value.
	if index < len(hl.Hosts) && hl.Hosts[index] == hostName {
		return true, index
	}

	return false, -1
}
