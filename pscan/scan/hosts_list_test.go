package scan_test

import (
	"errors"
	"os"
	"testing"

	"mnishiguchi.com/pscan/scan"
)

func TestAdd(t *testing.T) {
	testCases := []struct {
		name        string
		hostToAdd   string
		expectedLen int
		expectedErr error
	}{
		{"add new", "host2", 2, nil},
		{"add existing", "host1", 1, scan.ErrAlreadyExists},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hl := &scan.HostList{}

			// Initialize the list
			if err := hl.Add("host1"); err != nil {
				t.Fatal(err)
			}

			// Execute the Add method.
			err := hl.Add(tc.hostToAdd)

			// When some error is expected:
			if tc.expectedErr != nil {
				if err == nil {
					t.Fatalf("Expected error, got 'nil' instead\n")
				}

				if !errors.Is(err, tc.expectedErr) {
					t.Errorf("Expected error %q, got %q instead\n", tc.expectedErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got %q instead\n", err)
			}

			if len(hl.Hosts) != tc.expectedLen {
				t.Errorf("Expected list length %d, got %d instead\n", tc.expectedLen, len(hl.Hosts))
			}

			if hl.Hosts[1] != tc.hostToAdd {
				t.Errorf("Expected host name %q as index 1, got %q instead\n", tc.expectedLen, len(hl.Hosts))
			}
		})
	}
}

func TestRemove(t *testing.T) {
	testCases := []struct {
		name         string
		hostToRemove string
		expectedLen  int
		expectedErr  error
	}{
		{"remove existing", "host1", 1, nil},
		{"remove not found", "host3", 1, scan.ErrNotExists},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hl := &scan.HostList{}

			// Initialize the list
			for _, h := range []string{"host1", "host2"} {
				if err := hl.Add(h); err != nil {
					t.Fatal(err)
				}
			}

			// Execute the Remove method.
			err := hl.Remove(tc.hostToRemove)

			// When some error is expected:
			if tc.expectedErr != nil {
				if err == nil {
					t.Fatalf("Expected error, got nil instead\n")
				}

				if !errors.Is(err, tc.expectedErr) {
					t.Errorf("Expected error %q, got %q instead\n", tc.expectedErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got %q instead\n", err)
			}

			if len(hl.Hosts) != tc.expectedLen {
				t.Errorf("Expected list length %d, got %d instead\n", tc.expectedLen, len(hl.Hosts))
			}

			if hl.Hosts[0] == tc.hostToRemove {
				t.Errorf("Host name %q should not be in the list\n", tc.hostToRemove)
			}
		})
	}
}

func TestSaveLoad(t *testing.T) {
	hl1 := scan.HostList{} // For saving a list to a file
	hl2 := scan.HostList{} // For loading a list from a file

	// Initialize a list.
	hostName := "host1"
	hl1.Add(hostName)

	// Prepare a file that we save the list to.
	tempFile, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("Error creating temp file: %s", err)
	}
	defer os.Remove(tempFile.Name())

	// Save the list.
	if err := hl1.Save(tempFile.Name()); err != nil {
		t.Fatalf("Error saving list to file: %s", err)
	}

	// Load the list from the file.
	if err := hl2.Load(tempFile.Name()); err != nil {
		t.Fatalf("Error getting list from file: %s", err)
	}

	// The content of two lists should match.
	// Note that slices cannot be directly compared!
	if hl1.Hosts[0] != hl2.Hosts[0] {
		t.Errorf("Host %q should match %q host.", hl1.Hosts[0], hl2.Hosts[0])
	}
}

func TestLoadNoFile(t *testing.T) {
	// Prepare a file that we save the list to, then delete it so that the test
	// ensures that the file does not exist.
	temp, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("Error creating temp file: %s", err)
	}
	if err := os.Remove(temp.Name()); err != nil {
		t.Fatalf("Error deleting temp file: %s", err)
	}

	hl := &scan.HostList{}

	// Attempt to load a nonexistint file.
	if err := hl.Load(temp.Name()); err != nil {
		t.Errorf("Error no error, got %q instead\n", err)
	}
}
