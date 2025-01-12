package main

import (
    "testing"
    "time"
)


// Mock robots.txt data.
func mockRobotsTxtCrawlDelay(hostname string, delayMillis int) {
    rules := RobotsTxtRules{
        CrawlDelay: time.Duration(delayMillis) * time.Millisecond,
    }
    robotsRules[hostname] = rules
}

func TestCrawlDelay(t *testing.T) {
    hostname := "localhost"
    delayMillis := 100 // 100ms crawl delay.

    // Mock robots.txt with a 100ms crawl delay.
    mockRobotsTxtCrawlDelay(hostname, delayMillis)

    // Initial crawl attempt should be allowed.
    if shouldRespectCrawlDelay(hostname) {
        t.Errorf("Initial crawl attempt should be allowed for hostname: %s", hostname)
    }

    // Update the last crawl time.
    updateLastCrawlTime(hostname)

    // Next crawl attempt should be blocked if made before 100ms.
    if !shouldRespectCrawlDelay(hostname) {
        t.Errorf("Crawl attempt should be blocked if made before %dms for hostname: %s", delayMillis, hostname)
    }

    // Wait for the crawl delay duration.
    time.Sleep(time.Duration(delayMillis) * time.Millisecond)

    // Next crawl attempt should now be allowed.
    if shouldRespectCrawlDelay(hostname) {
        t.Errorf("Crawl attempt should be allowed after %dms for hostname: %s", delayMillis, hostname)
    }
}


// Mock robots.txt data.
func mockRobotsTxtDisallow(hostname string, disallowPaths []string) {
    rules := RobotsTxtRules{
        Disallow: disallowPaths,
    }
    robotsRules[hostname] = rules
}

func TestDisallow(t *testing.T) {
    hostname := "localhost"
    
    // Mock robots.txt with disallow patterns.
    mockRobotsTxtDisallow(hostname, []string{
        "/secret",
        "/private/*",
        "/chap21.html",
    })

    // URLs that should be disallowed.
    disallowedURLs := []string{
        "http://localhost/secret",
        "http://localhost/private/file.txt",
        "http://localhost/chap21.html",
    }

    for _, url := range disallowedURLs {
        if !isDisallowed(url, hostname) {
            t.Errorf("URL should be disallowed by robots.txt: %s", url)
        }
    }

    // URLs that should be allowed.
    allowedURLs := []string{
        "http://localhost/index.html",
        "http://localhost/public/info.html",
        "http://localhost/chap20.html",
    }

    for _, url := range allowedURLs {
        if isDisallowed(url, hostname) {
            t.Errorf("URL should not be disallowed by robots.txt: %s", url)
        }
    }
}
