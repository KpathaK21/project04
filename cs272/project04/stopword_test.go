package main

import (
	"encoding/json"
	"os"
	"testing"
//	"fmt"
)

// FetchStopWordsFromFile simulates fetching stop words from a local JSON file for testing purposes.
func FetchStopWordsFromFile(filePath string) map[string]struct{} {
	// Open the JSON file
	file, err := os.Open(filePath)
	if err != nil {
		panic("Unable to open stopwords.json file: " + err.Error())
	}
	defer file.Close()

	// Decode the JSON data into a slice of strings
	var stopWords []string
	if err := json.NewDecoder(file).Decode(&stopWords); err != nil {
		panic("Error decoding JSON stopwords: " + err.Error())
	}

	// Convert the slice to a map[string]struct{} for quick lookup
	stopWordMap := make(map[string]struct{})
	for _, word := range stopWords {
		stopWordMap[word] = struct{}{}
	}
	return stopWordMap
}

func TestStop(t *testing.T) {
    // Fetch stop words from the local JSON file.
    stopWords := FetchStopWordsFromFile("stopwords.json")

    // Initialize StopWordProcessor with the fetched stop words.
    swp := StopWordProcessor{
        stopWords: stopWords,
    }

    // Input: "this is a simple test"
    words := []string{"this", "is", "a", "simple", "test"}

    // Call RemoveStopWords to filter out the stop words.
    result := swp.RemoveStopWords(words)

    // Print the filtered result for debugging.
    t.Logf("Filtered result: %v", result)

    // Expected output: ["simple", "test"]
    expected := []string{"simple" , "test"}

    // Check if the length of the result matches the expected length.
    if len(result) != len(expected) {
        t.Fatalf("Expected length %d, got %d", len(expected), len(result))
    }

    // Loop through the result and expected output, comparing each word.
    for i, word := range expected {
        if result[i] != word {
            t.Errorf("Expected word %s, got %s", word, result[i])
        }
    }
}
