package service

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Analyze result - hold all the extected information
type AnalyzeResult struct {
	HTMLVersion string         `json:"htmlVersion"`
	PageTitle   string         `json:"pageTitle"`
	Headings    map[string]int `json:"headings"`
}

// Performs analysis on the given URL
func AnalyzeWebPage(targetURL string) (*AnalyzeResult, error) {
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = "https://" + targetURL
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(targetURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, errors.New("HTTP error: " + resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content := string(bodyBytes)

	result := &AnalyzeResult{}

	// Detect HTML Version
	result.HTMLVersion = detectHTMLVersion(content)

	// Extract Paget Titlt
	result.PageTitle = extractPageTitle(content)

	// Count Headings
	result.Headings = countHeadings(content)

	return result, nil

}

func detectHTMLVersion(content string) string {

	lowerContent := strings.ToLower(content)
	switch {
	case strings.Contains(lowerContent, "<!doctype html>"):
		return "HTML5"
	case strings.Contains(lowerContent, "xhtml 1.0"):
		return "XHTML 1.0"
	case strings.Contains(lowerContent, "xhtml 1.1"):
		return "XHTML 1.1"
	case strings.Contains(lowerContent, "html 4.01"):
		return "HTML 4.01"
	case strings.Contains(lowerContent, "html 3.2"):
		return "HTML 3.2"
	case strings.Contains(lowerContent, "html 2.0"):
		return "HTML 2.0"
	default:
		return "Unknown or custom HTML version"
	}
}

func extractPageTitle(content string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(content))
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return "No title found"
		case html.StartTagToken:
			t := tokenizer.Token()
			if t.Data == "title" {
				tokenizer.Next()
				return strings.TrimSpace(tokenizer.Token().Data)
			}
		}
	}
}

func countHeadings(content string) map[string]int {
	result := map[string]int{
		"h1": 0,
		"h2": 0,
		"h3": 0,
		"h4": 0,
		"h5": 0,
		"h6": 0,
	}

	tokenizer := html.NewTokenizer(strings.NewReader(content))
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			break
		}
		if tokenType == html.StartTagToken {
			t := tokenizer.Token()
			switch t.Data {
			case "h1", "h2", "h3", "h4", "h5", "h6":
				result[t.Data]++
			}
		}
	}
	return result
}
