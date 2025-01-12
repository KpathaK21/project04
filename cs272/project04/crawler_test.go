package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "strings"
    "io"
    "time"
    "fmt"
)

// TestDownloadWorker verifies the downloadWorker-like function retrieves content correctly.
func TestDownloadWorker(t *testing.T) {
    // Set up a mock server to simulate a webpage response.
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html") // Set Content-Type to text/html
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("<html><body>Hello, World!</body></html>"))
    })
    server := httptest.NewServer(handler)
    defer server.Close()

    // Define a function similar to downloadWorker for testing
    downloadContent := func(url string) ([]byte, error) {
        client := http.Client{Timeout: 30 * time.Second}
        for i := 0; i < 3; i++ {
            resp, err := client.Get(url)
            if err == nil {
                defer resp.Body.Close()
                if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/") {
                    return io.ReadAll(resp.Body)
                }
                return nil, fmt.Errorf("unsupported content type")
            }
            time.Sleep(2 * time.Second)
        }
        return nil, fmt.Errorf("failed to download after retries")
    }

    // Run the downloadContent function
    result, err := downloadContent(server.URL)
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }

    // Check the content returned
    expected := "<html><body>Hello, World!</body></html>"
    if strings.TrimSpace(string(result)) != expected {
        t.Errorf("Expected %s, got %s", expected, string(result))
    }
}
