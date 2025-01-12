package main


import (
	"reflect"
	"testing"

)


// TestExtract is a unit test for the Extract function
// It checks whether the function correctly extracts both words and hrefs from HTML content
func TestExtract(t *testing.T) {
	htmlData := []byte(`
	<!DOCTYPE html>
	<html>
	    <head>
	        <title>CS272 | Welcome</title>
	    </head>
	    <body>
	        <p>Hello World!</p>
	        <p>Welcome to <a href="https://cs272-f24.github.io/">CS272</a>!</p>
	    </body>
	</html>`)

	// The expected words that should be extracted from the HTML content
	expectedWords := []string{"Hello", "World!", "Welcome", "to", "CS272"}
	// The expected hrefs that should be extracted from the HTML <a> tags
	expectedHrefs := []string{"https://cs272-f24.github.io/"}

	// Call the Extract function to extract words and hrefs from the sample HTML data
	words, hrefs := Extract_html(htmlData)

	// Compare the extracted words with the expected result using reflect.DeepEqual
	// If they don't match, report an error.
	if !reflect.DeepEqual(words, expectedWords) {
		t.Errorf("Expected words %v, got %v", expectedWords, words)
	}

	// Compare the extracted hrefs with the expected result
	// If they don't match, report an error
	if !reflect.DeepEqual(hrefs, expectedHrefs) {
		t.Errorf("Expected hrefs %v, got %v", expectedHrefs, hrefs)
	}
}

// TestCleanHref tests the Clean function, which resolves relative URLs based on a base URL
func TestCleanHref(t *testing.T) {
	// Define the base URL to which relative hrefs will be resolved
	baseURL := "https://cs272-f24.github.io/"
	// The list of hrefs (some relative, some absolute) to be cleaned
	hrefs := []string{"/", "/help/", "/syllabus/", "https://gobyexample.com/"}

	// The expected output after resolving each href relative to the base URL
	expected := []string{
		"https://cs272-f24.github.io/",		// Relative URL "/" should resolve to base URL
		"https://cs272-f24.github.io/help/",	// "/help/" resolves to an absolute path
		"https://cs272-f24.github.io/syllabus/",	// "/syllabus/" resolves to an absolute path
		"https://gobyexample.com/",					// Absolute URLs should remain unchanged
	}

	// Loop through the list of hrefs and compare the cleaned result with the expected value
	for i, href := range hrefs {
		// Clean the current href and compare it with the expected result
		if cleaned := Clean(baseURL, href); cleaned != expected[i] {
			// Report an error if the cleaned URL does not match the expected value
			t.Errorf("Expected %s, got %s", expected[i], cleaned)
		}
	}
}
