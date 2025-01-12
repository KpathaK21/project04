package main

import (
    "fmt"
    "io"
    "time"
    "net/http"
    "strings"
    "regexp"
    "sync"
    "github.com/kljensen/snowball"
    "net/url"
)

// Crawler represents the main structure of the web crawler 
type Crawler struct {
    visited        map[string]bool       // Tracks URLs that have been visited to prevent re-crawling
    discovered     map[string]bool       // Tracks URLs that have been discovered and need indexing
    mu             sync.Mutex            // Mutex to handle concurrent map access for thread-safety
    urlChan        chan string           // Channel for managing URL processing tasks
    dataChan       chan downloadData     // Channel to send download results to processDownload
    clickableLinks map[string]string     // Stores discovered clickable links with titles
    wg             sync.WaitGroup        // WaitGroup to synchronize goroutine completion
    storage        *SQLiteStorage        // Storage interface for saving URL and term frequency data
}

// downloadData holds the results from the download function for each URL
type downloadData struct {
    url   string
    body  []byte
    error error
}

// NewCrawler initializes a new Crawler instance with the provided SQLiteStorage
func NewCrawler(storage *SQLiteStorage) *Crawler {
    return &Crawler{
        visited:        make(map[string]bool),
        discovered:     make(map[string]bool),
        urlChan:        make(chan string, 2000), // Buffered channel with capacity for efficient processing
        dataChan:       make(chan downloadData, 2000), // Channel to communicate downloaded data to processing
        clickableLinks: make(map[string]string),
        storage:        storage,
    }
}

// download fetches the HTML content of a given URL and sends the result to dataChan
func (c *Crawler) download(urlStr string) {
    // Parse the URL to obtain hostname information
    parsedURL, err := url.Parse(urlStr)
    if err != nil {
        fmt.Printf("Error parsing URL %s: %v\n", urlStr, err)
        return
    }
    hostname := parsedURL.Hostname()

    // Respect the crawl delay for the hostname if specified in robots.txt
    if shouldRespectCrawlDelay(hostname) {
        delay := robotsRules[hostname].CrawlDelay
        fmt.Printf("Waiting for crawl delay: %s\n", delay)
        time.Sleep(delay)
    }
    updateLastCrawlTime(hostname) // Update last crawl time after waiting

    client := http.Client{Timeout: 30 * time.Second} // Set HTTP client timeout for download
    resp, err := client.Get(urlStr)
    if err != nil {
        c.dataChan <- downloadData{url: urlStr, error: err}
        return
    }
    defer resp.Body.Close() // Ensure the response body is closed after reading

    contentType := resp.Header.Get("Content-Type")
    // Skip non-text content types to avoid downloading non-HTML files
    if !strings.HasPrefix(contentType, "text/") {
        c.dataChan <- downloadData{url: urlStr, error: fmt.Errorf("unsupported content type: %s", contentType)}
        return
    }

    // Read the response body and send the downloaded data to dataChan
    body, err := io.ReadAll(resp.Body)
    c.dataChan <- downloadData{url: urlStr, body: body, error: err}
}

// processDownload processes downloaded HTML content, extracts links, and updates term frequencies
func (c *Crawler) processDownload() {
    for data := range c.dataChan {
        if data.error != nil { // Check for download errors
            fmt.Printf("Error downloading URL %s: %v\n", data.url, data.error)
            c.wg.Done()
            continue
        }

        // Extract title and term frequencies from the HTML body
        title, termFreq := Extract_crawl(data.body)
        if err := c.storage.AddLink(data.url, title); err != nil {
            fmt.Printf("Error saving link: %v\n", err)
        }
        if err := c.storage.AddDocument(data.url, termFreq); err != nil {
            fmt.Printf("Error saving term frequencies: %v\n", err)
        }

        // Store the clickable link with its title
        c.mu.Lock()
        c.clickableLinks[data.url] = title
        c.mu.Unlock()

        // Extract hyperlinks from the downloaded HTML body
        _, hrefs := extractLinks(data.body)
        for _, href := range hrefs {
            cleanedHref := Clean(data.url, href) // Clean relative links based on base URL
            if strings.HasPrefix(cleanedHref, "https://www.usfca.edu") { // Filter for specific domain
                normalizedHref, err := NormalizeURL(cleanedHref) // Normalize the URL format
                
                if err == nil && !c.isVisited(normalizedHref) && !SkipRepeatingURL(normalizedHref) && !SkipLongPath(normalizedHref) {
                    // Parse hostname from URL for robots.txt compliance check
                    parsedURL, err := url.Parse(normalizedHref)
                    if err == nil {
                        hostname := parsedURL.Hostname()
                        
                        // Only process URL if it is not disallowed by robots.txt
                        if !isDisallowed(normalizedHref, hostname) {
                            c.markVisited(normalizedHref) // Mark URL as visited
                            c.wg.Add(1) // Increment WaitGroup counter
                            c.urlChan <- normalizedHref // Send URL to urlChan for further crawling
                        }
                    }
                }
            }
        }
        c.wg.Done() // Signal that processing for this URL is done
    }
}

// markVisited marks a URL as visited in a thread-safe manner to prevent duplicate processing
func (c *Crawler) markVisited(url string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.visited[url] = true
}

// isVisited checks if a URL has already been visited in a thread-safe manner
func (c *Crawler) isVisited(url string) bool {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.visited[url]
}

// startCrawling initializes the crawling process with download and processDownload goroutines
func (c *Crawler) startCrawling(baseURL string, numWorkers int) {
    c.wg.Add(1) // Initialize with one task in the WaitGroup

    // Close channels after all URLs have been processed
    go func() {
        c.wg.Wait()
        close(c.urlChan)
        close(c.dataChan)
    }()

    // Start worker goroutines for download and processDownload tasks
    for i := 0; i < numWorkers; i++ {
        go func() {
            for url := range c.urlChan {
                c.download(url) // Download content for each URL in urlChan
            }
        }()
        go c.processDownload() // Process downloaded content in dataChan
    }

    // Normalize the base URL before adding it to urlChan
    normalizedBaseURL, err := NormalizeURL(baseURL)
    if err != nil {
        fmt.Printf("Error normalizing base URL %s: %v\n", baseURL, err)
        return
    }

    c.urlChan <- normalizedBaseURL // Start crawling from the base URL
    c.discovered[normalizedBaseURL] = true
    c.wg.Wait() // Wait for all tasks to complete before closing

    fmt.Println("All unique links found and indexed")
}

// Extract_crawl extracts the title and text from the HTML body, normalizes and stems words,
// then calculates term frequencies for indexing
func Extract_crawl(body []byte) (string, map[string]int) {
    text := string(body)
    termFreq := make(map[string]int)

    titleRegex := regexp.MustCompile(`(?i)<title>(.*?)</title>`) // Regular expression to extract title
    matches := titleRegex.FindStringSubmatch(text)
    title := "Untitled" // Default title if none found
    if len(matches) > 1 {
        title = strings.TrimSpace(matches[1])
    }

    // Normalize and stem each word in the body text, adding it to the term frequency map
    words := strings.Fields(text)
    for _, word := range words {
        word = strings.ToLower(strings.Trim(word, ",.!?\"'();:")) // Remove punctuation and convert to lowercase
        if word != "" {
            stemmedWord, err := snowball.Stem(word, "english", true) // Stem the word
            if err == nil {
                termFreq[stemmedWord]++ // Increment frequency count for each stemmed word
            }
        }
    }
    return title, termFreq
}

// extractLinks extracts the hyperlinks (hrefs) from the HTML body
func extractLinks(body []byte) (string, []string) {
    text := string(body)
    hrefRegex := regexp.MustCompile(`(?i)<a[^>]+href="([^"]+)"`) // Regex to find href attributes in anchor tags
    hrefs := hrefRegex.FindAllStringSubmatch(text, -1)

    var links []string
    for _, href := range hrefs {
        if len(href) > 1 {
            links = append(links, href[1]) // Add each link to the list
        }
    }
    return text, links
}
