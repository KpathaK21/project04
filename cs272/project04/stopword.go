package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// StopWordProcessor holds a set of stop words for filtering
type StopWordProcessor struct {
	stopWords map[string]struct{} // A map for quick lookup of stop words
}

// NewStopWordProcessor initializes a StopWordProcessor by fetching stop words from a URL
func NewStopWordProcessor(url string) (*StopWordProcessor, error) {
	stopWords, err := FetchStopWords(url) // Fetch stop words from the provided URL
	if err != nil { // Check if there was an error fetching stop words
		return nil, err // Return the error if fetching fails
	}

	return &StopWordProcessor{
		stopWords: stopWords, // Assign the fetched stop words to the processor
	}, nil
}

// FetchStopWords loads stop words from a provided URL
// It returns a map[string]struct{} for O(1) lookup
func FetchStopWords(url string) (map[string]struct{}, error) {
	resp, err := http.Get(url) // Make a GET request to the URL to fetch stop words
	if err != nil { // Check if there was an error in making the request
		return nil, fmt.Errorf("error fetching stop words: %w", err) // Wrap and return the error
	}
	defer resp.Body.Close() // Ensure the response body is closed after reading

	// Decodes the JSON response into a slice of strings
	var stopWords []string // Declare a slice to hold the stop words
	if err := json.NewDecoder(resp.Body).Decode(&stopWords); err != nil { // Decode the JSON response
		return nil, fmt.Errorf("error decoding stop words: %w", err) // Return an error if decoding fails
	}

	// Convert the slice into a map for quick lookup of stop words
	stopWordMap := make(map[string]struct{}) // Create a map for O(1) lookup
	for _, word := range stopWords { // Iterate over the slice of stop words
		stopWordMap[word] = struct{}{} // Add each word to the map with an empty struct as the value
	}
	return stopWordMap, nil // Return the map of stop words
}

// RemoveStopWords filters out stop words from the input list of words
func (swp *StopWordProcessor) RemoveStopWords(words []string) []string {
	var filtered []string // Initialize a slice to store filtered words
	for _, word := range words { // Iterate over each word in the input list
		if _, found := swp.stopWords[word]; !found { // Check if the word is not in the stop words map
			filtered = append(filtered, word) // Append non-stop words to the filtered slice
		}
	}
	return filtered // Return the filtered list of words
}
