package service

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"web-analyzer/cache"

	"golang.org/x/net/html"
)

var linkCache = cache.NewLinkCache()

// Analyze result - hold all the expected information
type AnalyzeResult struct {
	HTMLVersion       string         `json:"htmlVersion"`
	PageTitle         string         `json:"pageTitle"`
	Headings          map[string]int `json:"headings"`
	Links             LinkAnalysis   `json:"links"`
	ContainsLoginForm bool           `json:"containsLoginForm"`
}

type LinkAnalysis struct {
	Total            int      `json:"total"`
	Internal         int      `json:"internal"`
	External         int      `json:"external"`
	Inaccessible     int      `json:"inaccessible"`
	InaccessibleURLs []string `json:"inaccessibleURLs,omitempty"`
}

// Performs analysis on the given URL
func AnalyzeWebPage(targetURL string) (*AnalyzeResult, error) {
	log.Printf("Info: Starting web page analysis for initial URL: %s\n", targetURL)
	if !strings.HasPrefix(targetURL, "http") {
		log.Printf("Info: URL missing protocol. adding https:// to %s\n", targetURL)
		targetURL = "https://" + targetURL
	}

	client := &http.Client{Timeout: 10 * time.Second}
	log.Printf("Info: Fetch content from %s with 10s timeout.\n", targetURL)
	resp, err := client.Get(targetURL)
	if err != nil {
		log.Printf("Error: Failed to fetch URL %s\n. Network error : %v\n", targetURL, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorMessage := fmt.Sprintf("HTTP error: Received status %s for URL %s", resp.Status, targetURL)
		return nil, errors.New(errorMessage)
	}
	log.Printf("Info: Successfully fetched content (Status: %s)", resp.Status)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		errorMessage := fmt.Sprintf("Error: Failed to read response body from %s: %v\n", targetURL, err)
		return nil, errors.New(errorMessage)
	}

	content := string(bodyBytes)
	log.Printf("Info: Content body read successfully for %s. Total size :%v", targetURL, len(content))

	result := &AnalyzeResult{}

	// detection HTML Version
	result.HTMLVersion = DetectHTMLVersion(targetURL, content)

	// extraction Paget Title
	result.PageTitle = ExtractPageTitle(targetURL, content)

	// Count Headings
	result.Headings = CountHeadings(targetURL, content)

	// Analyze Links
	result.Links = AnalyzeLinks(targetURL, content)

	// Check Login Form
	result.ContainsLoginForm = CheckForLoginForm(content)

	return result, nil

}

func DetectHTMLVersion(targetURL, content string) string {
	log.Printf("Info: Starting detect html version for base URL : %s\n", targetURL)
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

func ExtractPageTitle(targetURL, content string) string {
	log.Printf("Info: Starting extract page title for base URL: %s\n", targetURL)
	tokenizer := html.NewTokenizer(strings.NewReader(content))
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return "No title found"
		case html.StartTagToken:
			t := tokenizer.Token()
			if t.Data == "title" {
				if tokenizer.Next() == html.TextToken {
					title := strings.TrimSpace(tokenizer.Token().Data)
					log.Printf("Info: Finished extract title from %s: %s\n", targetURL, title)
					return title
				}
			}
		}
	}
}

func CountHeadings(targetURL, content string) map[string]int {
	log.Printf("Info: Starting count headers for base URL: %s\n", targetURL)
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
	log.Printf("Info: Finished counting headers for %s: %v\n", targetURL, result)
	return result
}

func AnalyzeLinks(baseURL, content string) LinkAnalysis {
	log.Printf("Info: Starting link analysis for base URL: %s\n", baseURL)

	tokenizer := html.NewTokenizer(strings.NewReader(content))
	var links []string

	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			if tokenizer.Err() != nil && tokenizer.Err().Error() != "EOF" {
				log.Printf("Warning: HTML tokenizing error encountered: %v\n", tokenizer.Err())
			}
			break
		}

		if tokenType == html.StartTagToken {
			token := tokenizer.Token()
			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" && attr.Val != "" {
						links = append(links, attr.Val)
					}
				}
			}
		}
	}

	log.Printf("Info: Finished HTML parsing. Found %d potential links.\n", len(links))

	result := LinkAnalysis{
		Total: len(links),
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		log.Fatalf("Error: Parsing provided baseURL %q: %v\n", baseURL, err)
	}

	client := &http.Client{Timeout: 3 * time.Second}

	var mutx sync.Mutex
	var wGrp sync.WaitGroup

	concurrencyLimit := 10
	sem := make(chan struct{}, concurrencyLimit)

	for i, link := range links {
		wGrp.Add(1)
		go func(i int, link string) {
			defer wGrp.Done()
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			log.Printf("Info: Processing link %d/%d: %s\n", i+1, len(links), link)

			ul, err := url.Parse(link)
			if err != nil {
				mutx.Lock()
				result.Inaccessible++
				result.InaccessibleURLs = append(result.InaccessibleURLs, link)
				mutx.Unlock()
				return
			}

			if !ul.IsAbs() {
				ul = base.ResolveReference(ul)
			}

			absoluteURL := ul.String()

			if cachedAccessible, found := linkCache.Get(absoluteURL); found {
				mutx.Lock()
				if cachedAccessible {
					if ul.Host == base.Host {
						result.Internal++
					} else {
						result.External++
					}
				} else {
					result.Inaccessible++
					result.InaccessibleURLs = append(result.InaccessibleURLs, absoluteURL)
				}
				mutx.Unlock()
				log.Printf("Info: Cache hit for %s (accessible: %t)\n", absoluteURL, cachedAccessible)
				return
			}

			resp, err := client.Head(absoluteURL)
			accessible := (err != nil || resp.StatusCode >= 400)
			if resp != nil {
				resp.Body.Close()
			}

			linkCache.Set(absoluteURL, accessible)

			mutx.Lock()
			if !accessible {
				result.Inaccessible++
				result.InaccessibleURLs = append(result.InaccessibleURLs, absoluteURL)
			} else if ul.Host == base.Host {
				result.Internal++
			} else {
				result.External++
			}
			mutx.Unlock()

		}(i, link)
	}

	wGrp.Wait()

	log.Printf("Info: Analysis complete. Total: %d, Internal: %d, External: %d, Inaccessible: %d.\n",
		result.Total, result.Internal, result.External, result.Inaccessible)

	return result
}

// Using simple string search approach for find login form.
func CheckForLoginForm(content string) bool {
	log.Printf("Info: Starting login form detection")

	// Simple string search - most reliable for basic detection
	hasForm := strings.Contains(strings.ToLower(content), "<form")
	hasPassword := strings.Contains(strings.ToLower(content), `type="password"`)

	result := hasForm && hasPassword
	log.Printf("Info: Login form detection - Form: %t, Password: %t, Result: %t",
		hasForm, hasPassword, result)

	return result
}
