package main

import (
    "regexp"
    "strconv"
    "strings"
    "time"
)

// RobotsTxtRules stores disallow and crawl-delay rules for a particular user agent.
type RobotsTxtRules struct {
    UserAgent   string         // User agent string for the rules.
    Disallow    []string       // List of disallowed paths.
    CrawlDelay  time.Duration  // Duration the crawler should wait between requests.
}

// Store rules for each hostname.
var robotsRules = make(map[string]RobotsTxtRules) // Map to hold robots.txt rules for different hostnames.

// Store last crawl time for each hostname.
var lastCrawlTime = make(map[string]time.Time) // Map to track the last time the crawler accessed each hostname.

// parseRobotsTxt parses the robots.txt content and extracts Disallow and Crawl-delay rules.
func parseRobotsTxt(hostname string, robotsTxtContent string) {
    rules := RobotsTxtRules{
        CrawlDelay: 100 * time.Millisecond, // Default Crawl-delay is 500ms.
    }

    // Regular expressions to match Disallow and Crawl-delay lines.
    disallowRegex := regexp.MustCompile(`(?i)^Disallow:\s*(.*)`)  // Case-insensitive match for Disallow.
    crawlDelayRegex := regexp.MustCompile(`(?i)^Crawl-delay:\s*(\d+)`) // Case-insensitive match for Crawl-delay.

    lines := strings.Split(robotsTxtContent, "\n") // Split content into lines.
    for _, line := range lines {
        line = strings.TrimSpace(line) // Remove leading and trailing whitespace.

        // Check for Disallow records.
        if disallowMatch := disallowRegex.FindStringSubmatch(line); len(disallowMatch) > 1 {
            disallowPath := disallowMatch[1] // Extract the disallowed path.
            rules.Disallow = append(rules.Disallow, disallowPath) // Append to the Disallow list.
        }

        // Check for Crawl-delay.
        if crawlDelayMatch := crawlDelayRegex.FindStringSubmatch(line); len(crawlDelayMatch) > 1 {
            delay, err := strconv.Atoi(crawlDelayMatch[1]) // Convert delay string to integer.
            if err == nil {
                rules.CrawlDelay = time.Duration(delay) * time.Millisecond // Set CrawlDelay if parsing is successful.
            }
        }
    }

    // Store the parsed rules for the given hostname.
    robotsRules[hostname] = rules // Save the rules in the global map.
}

// isDisallowed checks if a URL is disallowed for crawling based on robots.txt rules.
func isDisallowed(url string, hostname string) bool {
    rules, exists := robotsRules[hostname] // Retrieve rules for the hostname.
    if !exists {
        return false // No rules, allow crawling by default.
    }

    // Check if the URL matches any disallowed patterns.
    for _, disallow := range rules.Disallow {
        matched, err := regexp.MatchString(disallow, url) // Check if the URL matches the disallow rule.
        if err == nil && matched {
            return true // URL is disallowed if matched.
        }
    }
    return false // URL is allowed if no disallow rules match.
}

// shouldRespectCrawlDelay checks if the crawler should wait before making the next request.
func shouldRespectCrawlDelay(hostname string) bool {
    rules, exists := robotsRules[hostname]
    if !exists {
        return false // No crawl delay rules, proceed with the request.
    }

    lastTime, timeExists := lastCrawlTime[hostname]
    if !timeExists {
        updateLastCrawlTime(hostname) // Allow and set the time for the first crawl.
        return false
    }

    if time.Since(lastTime) >= rules.CrawlDelay {
        updateLastCrawlTime(hostname) // Allow and reset the crawl timer.
        return false
    }

    return true // Block if within the crawl delay.
}

// updateLastCrawlTime updates the last crawl time for a given hostname.
func updateLastCrawlTime(hostname string) {
    lastCrawlTime[hostname] = time.Now() // Set the current time as the last crawl time for the hostname.
}
