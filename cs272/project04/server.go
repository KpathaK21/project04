package main

import (
    "fmt"
    "net/http"
    "sync"
    "math"
    "sort"
    "io"
    "os"
)

// Server struct defines the web server and stores a reference to the storage layer
type Server struct {
    storage   *SQLiteStorage   // Storage interface for retrieving document and term frequency data
    mu        sync.Mutex       // Mutex for managing concurrent access to storage
}

// Hit struct represents a search result hit, storing the URL and its calculated TF-IDF score
type Hit struct {
    URL   string
    Score float64 
    Title string
}

// NewServer initializes and returns a new Server instance with the given storage
func NewServer(storage *SQLiteStorage) *Server {
    return &Server{
        storage: storage,
    }
}

// CalculateTFIDF calculates the TF-IDF scores for a given term across documents.
func (s *Server) TFIDF(term string) ([]Hit, error) {
    var results []Hit

    // Retrieve the total number of documents in storage to calculate IDF
    docCount, err := s.storage.GetDocumentCount()
    if err != nil || docCount == 0 {
        return nil, fmt.Errorf("no documents found in the storage")
    }

    // Retrieve the document frequency (DF) for the search term
    df, err := s.storage.GetDocumentFrequency(term)
    if err != nil || df == 0 {
        return nil, fmt.Errorf("no documents contain the term '%s'", term)
    }

    // Calculate the Inverse Document Frequency (IDF) for the term
    idf := math.Log(float64(docCount) / float64(df))

    // Retrieve the term frequency (TF) for the term across all documents
    termFreq, err := s.storage.GetTermFrequency(term)
    if err != nil || len(termFreq) == 0 {
        return nil, fmt.Errorf("no term frequencies found for '%s'", term)
    }

    // Calculate the TF-IDF score for each document
    for url, freq := range termFreq {
        tf := float64(freq)
        score := tf * idf

        // Retrieve the title for each URL
        title, err := s.storage.GetTitle(url)
        if err != nil {
            title = "Untitled" // Fallback title if not found
        }

        results = append(results, Hit{URL: url, Title: title, Score: score})
    }

    return results, nil
}

// search performs a search for a given term and retrieves the sorted results by TF-IDF score.
func (s *Server) search(term string) []Hit {
    results, err := s.TFIDF(term)
    if err != nil {
        fmt.Println(err)
        return nil
    }

    // Sort results by score in descending order
    sort.Slice(results, func(i, j int) bool {
        return results[i].Score > results[j].Score
    })

    return results
}

// // search performs a search using the TF-IDF algorithm for a given term and retrieves document titles
// func (s *Server) search(term string) []Hit {
//     var results []Hit 
// 
//     // Retrieve the total number of documents in storage to calculate IDF
//     docCount, err := s.storage.GetDocumentCount()
//     if err != nil || docCount == 0 {
//         fmt.Println("No documents found in the storage")
//         return results // Return empty results if no documents are found
//     }
//     fmt.Printf("Document Count: %d\n", docCount)
// 
//     // Retrieve the document frequency (DF) for the search term
//     df, err := s.storage.GetDocumentFrequency(term)
//     if err != nil || df == 0 {
//         fmt.Printf("No documents contain the term '%s'\n", term)
//         return results // Return empty results if the term is not found
//     }
//     fmt.Printf("Document Frequency for '%s': %d\n", term, df)
// 
//     // Calculate the Inverse Document Frequency (IDF) for the term
//     idf := math.Log(float64(docCount) / float64(df)) // IDF helps weight less common terms higher
//     fmt.Printf("IDF for '%s': %f\n", term, idf)
// 
//     // Retrieve the term frequency (TF) for the term across all documents
//     termFreq, err := s.storage.GetTermFrequency(term)
//     if err != nil || len(termFreq) == 0 {
//         fmt.Printf("No term frequencies found for '%s'\n", term)
//         return results
//     }
// 
//     // Calculate the TF-IDF score for each document containing the term and retrieve the title
//     for url, freq := range termFreq {
//         tf := float64(freq) // Convert term frequency to float for calculation
//         score := tf * idf // Calculate the TF-IDF score
// 
//         // Retrieve the title for each URL from storage
//         title, err := s.storage.GetTitle(url)
//         if err != nil {
//             title = "Untitled" // Fallback title if not found
//         }
// 
//         // Append the result with URL, title, and score to the results slice
//         results = append(results, Hit{URL: url, Title: title, Score: score})
//     }
// 
//     // Sort results by score in descending order for relevance
//     sort.Slice(results, func(i, j int) bool {
//         return results[i].Score > results[j].Score
//     })
// 
//     fmt.Printf("Found %d results for term '%s'\n", len(results), term)
//     return results
// }

// SearchHandler handles incoming HTTP requests for the /search endpoint
// It retrieves the search term from the query and returns relevant results as clickable links with titles
func (s *Server) SearchHandler(w http.ResponseWriter, r *http.Request) {
    searchTerm := r.URL.Query().Get("term") // Get search term from query parameter
    w.Header().Set("Content-Type", "text/html") // Set response content type to HTML

    if searchTerm == "" { // Check if the search term is provided
        fmt.Fprintln(w, "<p>Please provide a search term using the 'term' query parameter</p>")
        return
    }

    hits := s.search(searchTerm) // Perform search and get results

    if len(hits) == 0 { // Check if any results are found
        fmt.Fprintln(w, "<p>No results found for the search term</p>")
        return
    }

    // Display search results as clickable links with titles
    fmt.Fprintf(w, "<h1>Search results for: %s</h1>", searchTerm)
    for _, hit := range hits {
        fmt.Fprintf(w, "<p><a href='%s' target='_blank'>%s</a> - Score: %.6f</p>", hit.URL, hit.Title, hit.Score)
    }
}

// ClickableLinksHandler serves an HTML page displaying clickable links
func (s *Server) ClickableLinksHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html") // Set response content type to HTML
    fmt.Fprintln(w, "<h1>Clickable Links on usfca.edu</h1>")

    // Retrieve all links with titles from the storage
    links, err := s.storage.GetAllLinks()
    if err != nil {
        fmt.Fprintln(w, "Error retrieving links")
        return
    }

    // Display each link with its title as an HTML anchor tag
    for _, link := range links {
        fmt.Fprintf(w, "<a href='%s' target='_blank'>%s</a><br/>", link.URL, link.Title)
    }
}

// StartServer configures HTTP routes and starts the server to handle requests
func (s *Server) StartServer() {
    http.Handle("/", http.FileServer(http.Dir("static")))               // Serves static files from the "static" directory
    http.HandleFunc("/search", s.SearchHandler)                         // Route for search requests
    http.HandleFunc("/clickable-links", s.ClickableLinksHandler)        // Route for clickable links page

    fmt.Println("Starting server on :8080...")
    if err := http.ListenAndServe(":8080", nil); err != nil {           // Start the server on port 8080
        fmt.Printf("Failed to start server: %v\n", err)
    }
}

func main() {
    // Initialize SQLite storage and exit if there is an error
    storage, err := NewSQLiteStorage("search.db")
    if err != nil {
        fmt.Printf("Error initializing SQLite: %v\n", err)
        os.Exit(1)
    }

    server := NewServer(storage)   // Initialize the server with the storage layer
    crawler := NewCrawler(storage) // Initialize the crawler with the same storage layer

    // Start the crawler as a background goroutine
    go func() {
        // Retrieve and parse robots.txt for crawling restrictions
        robotsURL := "https://www.usfca.edu/robots.txt"
        resp, err := http.Get(robotsURL)
        if err == nil {
            defer resp.Body.Close()
            robotsTxtContent, err := io.ReadAll(resp.Body)
            if err != nil {
                fmt.Printf("Error reading robots.txt content: %v\n", err)
            } else {
                parseRobotsTxt("www.usfca.edu", string(robotsTxtContent)) // Parse robots.txt content
            }
        } else {
            fmt.Printf("Error fetching robots.txt: %v\n", err)
        }

        // Start crawling from the base URL using multiple workers
        crawler.startCrawling("https://www.usfca.edu", 5)
    }()

    // Start the HTTP server to handle incoming requests
    server.StartServer()
}
