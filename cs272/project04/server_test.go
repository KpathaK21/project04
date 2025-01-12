package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    "os"
    "math"
    
)

//setupTestStorage initializes a temporary SQLite storage for testing.
func setupTestStorage(t *testing.T) *SQLiteStorage {
    os.Remove("test.db") // Clean up the test database file

    storage, err := NewSQLiteStorage("test.db")
    if err != nil {
        t.Fatalf("Failed to initialize SQLite storage: %v", err)
    }

    return storage
}

// TestNewServer tests if the server initializes correctly.
func TestNewServer(t *testing.T) {
    storage := setupTestStorage(t)
    defer os.Remove("test.db")

    server := NewServer(storage)
    if server == nil {
        t.Fatalf("NewServer() returned nil")
    }
}

// // TestSearch tests the search functionality.
// func TestSearch(t *testing.T) {
//     storage := setupTestStorage(t)
//     defer os.Remove("test.db")
// 
//     server := NewServer(storage)
// 
//     // Insert mock data for testing search functionality
//     storage.AddDocument("example.com", map[string]int{"term": 2})
//     
//     result := server.search("term")
// 
//     expected := []Hit{
//         {URL: "example.com", Score: 0.0},
//     }
// 
//     if len(result) != len(expected) {
//         t.Errorf("Expected length %d but got %d", len(expected), len(result))
//     }
// 
//     for i, hit := range result {
//         if hit.URL != expected[i].URL || hit.Score != expected[i].Score {
//             t.Errorf("Expected %+v but got %+v", expected[i], hit)
//         }
//     }
// }

func TestSearch(t *testing.T) {
    // Setup a temporary test storage
    storage := setupTestStorage(t)
    defer os.Remove("test.db")

    server := NewServer(storage)

    // Insert mock data for testing
    err := storage.AddDocument("doc1.com", map[string]int{"search": 5})
    if err != nil {
        t.Fatalf("Failed to add document to storage: %v", err)
    }

    err = storage.AddDocument("doc2.com", map[string]int{"search": 3})
    if err != nil {
        t.Fatalf("Failed to add document to storage: %v", err)
    }

    err = storage.AddDocument("doc3.com", map[string]int{"example": 2})
    if err != nil {
        t.Fatalf("Failed to add document to storage: %v", err)
    }

    // Add titles for the documents
    err = storage.AddLink("doc1.com", "Document 1")
    if err != nil {
        t.Fatalf("Failed to add link to storage: %v", err)
    }

    err = storage.AddLink("doc2.com", "Document 2")
    if err != nil {
        t.Fatalf("Failed to add link to storage: %v", err)
    }

    err = storage.AddLink("doc3.com", "Document 3")
    if err != nil {
        t.Fatalf("Failed to add link to storage: %v", err)
    }

    // Call the search function
    hits := server.search("search")

    // Expected TF-IDF scores
    docCount := 3
    df := 2
    idf := math.Log(float64(docCount) / float64(df))
    expected := []Hit{
        {URL: "doc1.com", Title: "Document 1", Score: 5 * idf},
        {URL: "doc2.com", Title: "Document 2", Score: 3 * idf},
    }

    // Validate results
    if len(hits) != len(expected) {
        t.Errorf("Expected %d hits, got %d", len(expected), len(hits))
    }

    for i, hit := range hits {
        if hit.URL != expected[i].URL || hit.Title != expected[i].Title || hit.Score != expected[i].Score {
            t.Errorf("Expected %+v, got %+v", expected[i], hit)
        }
    }

    // Ensure the results are sorted by score in descending order
    for i := 1; i < len(hits); i++ {
        if hits[i-1].Score < hits[i].Score {
            t.Errorf("Results are not sorted by score in descending order: %+v", hits)
        }
    }
}


func TestTFIDF(t *testing.T) {
    // Setup a temporary test storage
    storage := setupTestStorage(t)
    defer os.Remove("test.db")

    server := NewServer(storage)

    // Insert mock data for testing
    err := storage.AddDocument("doc1.com", map[string]int{"test": 3})
    if err != nil {
        t.Fatalf("Failed to add document to storage: %v", err)
    }

    err = storage.AddDocument("doc2.com", map[string]int{"test": 2})
    if err != nil {
        t.Fatalf("Failed to add document to storage: %v", err)
    }

    err = storage.AddDocument("doc3.com", map[string]int{"example": 1})
    if err != nil {
        t.Fatalf("Failed to add document to storage: %v", err)
    }

    // Add titles for the documents
    err = storage.AddLink("doc1.com", "Document 1")
    if err != nil {
        t.Fatalf("Failed to add link to storage: %v", err)
    }

    err = storage.AddLink("doc2.com", "Document 2")
    if err != nil {
        t.Fatalf("Failed to add link to storage: %v", err)
    }

    err = storage.AddLink("doc3.com", "Document 3")
    if err != nil {
        t.Fatalf("Failed to add link to storage: %v", err)
    }

    // Call the TFIDF function
    hits, err := server.TFIDF("test")
    if err != nil {
        t.Fatalf("Unexpected error in TFIDF: %v", err)
    }

    // Expected TF-IDF scores
    docCount := 3
    df := 2
    idf := math.Log(float64(docCount) / float64(df))
    expected := []Hit{
        {URL: "doc1.com", Title: "Document 1", Score: 3 * idf},
        {URL: "doc2.com", Title: "Document 2", Score: 2 * idf},
    }

    // Validate results
    if len(hits) != len(expected) {
        t.Errorf("Expected %d hits, got %d", len(expected), len(hits))
    }

    for i, hit := range hits {
        if hit.URL != expected[i].URL || hit.Title != expected[i].Title || hit.Score != expected[i].Score {
            t.Errorf("Expected %+v, got %+v", expected[i], hit)
        }
    }
}


// TestSearchHandler tests the HTTP search handler.
func TestSearchHandler(t *testing.T) {
    storage := setupTestStorage(t)
    defer os.Remove("test.db")

    server := NewServer(storage)

    // Insert mock data to simulate search functionality
    storage.AddDocument("example.com", map[string]int{"romeo": 1})

    req := httptest.NewRequest("GET", "/search?term=romeo", nil)
    w := httptest.NewRecorder()

    server.SearchHandler(w, req)
    resp := w.Result()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status OK, got %v", resp.Status)
    }
}

// TestClickableLinksHandler tests the ClickableLinksHandler for serving links.
func TestClickableLinksHandler(t *testing.T) {
    storage := setupTestStorage(t)
    defer os.Remove("test.db")

    server := NewServer(storage)

    // Insert mock link data
    storage.AddLink("https://www.example.com", "Example Title")

    req := httptest.NewRequest("GET", "/clickable-links", nil)
    w := httptest.NewRecorder()

    server.ClickableLinksHandler(w, req)
    resp := w.Result()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status OK, got %v", resp.Status)
    }
}

// TestStartServer checks if the server can start without panicking.
func TestStartServer(t *testing.T) {
    storage := setupTestStorage(t)
    defer os.Remove("test.db")

    server := NewServer(storage)

    go func() {
        defer func() {
            if r := recover(); r != nil {
                t.Errorf("The server panicked: %v", r)
            }
        }()
        server.StartServer()
    }()

    time.Sleep(100 * time.Millisecond)
}
