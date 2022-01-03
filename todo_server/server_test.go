package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	testCases := []struct {
		name             string
		path             string
		expectedCode     int
		expectedNumItems int
		expectedContent  string
	}{
		{name: "GET /", path: "/", expectedCode: http.StatusOK, expectedContent: "There is an API here"},
		{name: "Not found", path: "/todo/500", expectedCode: http.StatusNotFound},
	}

	url, cleanup := setupAPI(t)
	defer cleanup()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				body []byte
				err  error
			)

			// Issue a GET request
			resp, err := http.Get(url + tc.path)
			if err != nil {
				t.Error(err)
			}
			defer resp.Body.Close()

			// Check the status code
			if resp.StatusCode != tc.expectedCode {
				t.Fatalf("Expected %q, got %q", http.StatusText(tc.expectedCode), http.StatusText(resp.StatusCode))
			}

			// Check the response content
			switch {
			case strings.Contains(resp.Header.Get("Content-Type"), "text/plain"):
				// Read the response body
				if body, err = io.ReadAll(resp.Body); err != nil {
					t.Error(err)
				}

				// Verify the response body is as expected
				if !strings.Contains(string(body), tc.expectedContent) {
					t.Errorf("Expected %q, got %q", tc.expectedContent, string(body))
				}

			default:
				t.Fatalf("Unsupported Content-Type: %q", resp.Header.Get("Content-Type"))
			}
		})
	}

}

func setupAPI(t *testing.T) (string, func()) {
	t.Helper()                                  // Mark this test as a test helper.
	s := httptest.NewServer(newMultiplexer("")) // Create a test server.

	return s.URL, func() { s.Close() }
}
