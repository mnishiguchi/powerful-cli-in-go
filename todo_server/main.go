package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	host := flag.String("h", "localhost", "Server host")
	port := flag.Int("p", 8080, "Server port")
	todoFile := flag.String("f", "todo_server.json", "todo JSON file")
	flag.Parse()

	// Instantiate an HTTP server specifying options.
	s := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", *host, *port),
		Handler:      newMultiplexer(*todoFile),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Listen for incoming requests.
	if err := s.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// We could control how the HTTP server behaves when a connection closes, or
	// we could gracefully stop it. We do not need to worry about these options
	// since we are using this server for testing and not for handling real workload.
}

func newMultiplexer(todoFile string) http.Handler {
	// The http.ServeMux type satisfies the http.Handler interface.
	m := http.NewServeMux()
	m.HandleFunc("/", rootHandler)

	return m
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}

	content := "There is an API here"
	replyTextContent(w, req, http.StatusOK, content)
}

func replyTextContent(w http.ResponseWriter, req *http.Request, status int, content string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(content))
}
