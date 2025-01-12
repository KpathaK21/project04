package main

import (
	"bytes"
	"net/url"
	"golang.org/x/net/html"
	"strings"
	"unicode"
	"log"
)

// Extract parses HTML content to extract words and hrefs.
// It returns two slices: one with words from text nodes and one with href attributes from <a> tags.
func Extract_html(htmlContent []byte) ([]string, []string) {
	var words, hrefs []string

	// Parse the HTML content into a document tree structure.
	doc, err := html.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		// If parsing fails, return empty slices.
		return words, hrefs
	}

	// Traverse the parsed document to extract words and hrefs.
	traverseNodes(doc, &words, &hrefs)

	return words, hrefs
}

// traverseNodes recursively traverses an HTML node tree to extract text and hrefs.
// It extracts words from text nodes and href attributes from <a> elements.
func traverseNodes(n *html.Node, words *[]string, hrefs *[]string) {
	// Begin extraction if the current node is the <body> tag.
	if n.Type == html.ElementNode && n.Data == "body" {
		extractFromBody(n, words, hrefs)
		return
	}

	// Recursively traverse child nodes.
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		traverseNodes(c, words, hrefs)
	}
}

// extractFromBody processes nodes within the <body> tag, extracting text and hrefs.
// Text nodes are passed to extractText, while <a> tags are passed to extractHref.
func extractFromBody(n *html.Node, words *[]string, hrefs *[]string) {
	if n.Type == html.TextNode {
		// Extract words from text nodes.
		extractText(n.Data, words)
	} else if n.Type == html.ElementNode && n.Data == "a" {
		// Extract href attributes from <a> tags.
		extractHref(n, hrefs)
	}

	// Recursively process child nodes within the <body>.
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractFromBody(c, words, hrefs)
	}
}

// extractText parses a text node into words and adds them to the words slice.
// It trims spaces, handles punctuation, and maintains word boundaries.
func extractText(text string, words *[]string) {
	// Trim leading and trailing whitespace from the text.
	cleanedText := strings.TrimSpace(text)

	// Use a string builder to construct individual words.
	var wordBuilder strings.Builder
	for i, r := range cleanedText {
		if unicode.IsSpace(r) {
			// Finalize the current word on encountering a space.
			if wordBuilder.Len() > 0 {
				*words = append(*words, wordBuilder.String())
				wordBuilder.Reset()
			}
		} else {
			// Handle punctuation by attaching it to the current word if not at the end.
			if unicode.IsPunct(r) && (i == len(cleanedText)-1 || unicode.IsSpace(rune(cleanedText[i+1]))) {
				if wordBuilder.Len() > 0 {
					wordBuilder.WriteRune(r)
				}
			} else {
				// Continue building the current word.
				wordBuilder.WriteRune(r)
			}
		}
	}

	// Append any remaining word in the builder to the words slice.
	if wordBuilder.Len() > 0 {
		*words = append(*words, wordBuilder.String())
	}
}

// extractHref searches for href attributes in <a> tags and adds them to the hrefs slice.
func extractHref(n *html.Node, hrefs *[]string) {
	for _, attr := range n.Attr {
		if attr.Key == "href" {
			// Add the href value to the hrefs slice.
			*hrefs = append(*hrefs, attr.Val)
		}
	}
}

// Clean resolves relative URLs to absolute URLs using a base URL.
// It parses and combines the base URL and href URL to produce a fully resolved URL.
func Clean(base string, href string) string {
	baseURL, err := url.Parse(base)
	if err != nil {
		// If base URL parsing fails, log the error and return the original href.
		log.Printf("Error parsing base URL '%s': %v", base, err)
		return href
	}

	hrefURL, err := url.Parse(href)
	if err != nil {
		// If href URL parsing fails, log the error and return the original href.
		log.Printf("Error parsing href URL '%s': %v", href, err)
		return href
	}

	// Return the absolute URL by resolving href relative to the base URL.
	return baseURL.ResolveReference(hrefURL).String()
}

// NormalizeURL standardizes URLs by removing fragments, queries, and enforcing trailing slashes.
func NormalizeURL(rawURL string) (string, error) {
    parsedURL, err := url.Parse(rawURL)
    if err != nil {
        return "", err
    }
    parsedURL.Fragment = "" // Remove URL fragment
    parsedURL.RawQuery = "" // Remove query parameters

    // Enforce trailing slash if absent.
    if parsedURL.Path == "" || parsedURL.Path[len(parsedURL.Path)-1] != '/' {
        parsedURL.Path += "/"
    }
    return strings.ToLower(parsedURL.String()), nil
}

// SkipRepetitiveURL checks for repeating segments in a URL path, which may indicate redundancy.
func SkipRepeatingURL(urlStr string) bool {
    parsedURL, err := url.Parse(urlStr)
    if err != nil {
        return false // Return false if URL parsing fails
    }
    segments := strings.Split(parsedURL.Path, "/")
    // Identify three consecutive identical segments in the path.
    for i := 0; i < len(segments)-2; i++ {
        if segments[i] != "" && segments[i] == segments[i+1] && segments[i+1] == segments[i+2] {
            return true
        }
    }
    return false
}

// SkipLongPath returns true if the URL path has more than a threshold number of segments.
func SkipLongPath(urlStr string) bool {
    parsedURL, err := url.Parse(urlStr)
    if err != nil {
        return false // Return false if URL parsing fails
    }
    segments := strings.Split(parsedURL.Path, "/")
    return len(segments) > 10 // Threshold set to 10 segments
}
