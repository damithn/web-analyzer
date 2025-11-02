package service

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Analyze result - hold all the extected information
type AnalyzeResult struct {
	HTMLVersion      string         `json:"htmlVersion"`
	PageTitle        string         `json:"pageTitle"`
	Headings         map[string]int `json:"headings"`
	Links            LinkAnalysis   `json:"links"`
	ContainLoginform bool           `json:"containLoginForm"`
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
	log.Printf("Starting web page analysis for initial URL: %s\n", targetURL)
	if !strings.HasPrefix(targetURL, "http") {
		log.Printf("Info: URL missing protocol. adding https:// to %s\n", targetURL)
		targetURL = "https://" + targetURL
	}

	client := &http.Client{Timeout: 10 * time.Second}
	log.Printf("Info: Fetch content from %s with 10s timeout.\n", targetURL)
	resp, err := client.Get(targetURL)
	if err != nil {
		log.Printf("Error: Failed to fetch URL %s\n. Netoerk error : %v\n", targetURL, err)
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
	log.Printf("Info: Content body read sussessfully for %s. Total size :%v", targetURL, len(content))

	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		log.Printf("Erro: faile to parse HTML content for %s: %v\n", targetURL, err)
		return nil, fmt.Errorf("failed to parse HTML: %s", err)
	}

	result := &AnalyzeResult{}

	// Detect HTML Version
	result.HTMLVersion = DetectHTMLVersion(targetURL, content)

	// Extract Paget Titlt
	result.PageTitle = ExtractPageTitle(targetURL, content)

	// Count Headings
	result.Headings = CountHeadings(targetURL, content)

	// Analysis Links
	result.Links = AnalyzeLinks(targetURL, content)

	// Check Login Form
	result.ContainLoginform = CheckForLoginForm(doc)

	return result, nil

}

func DetectHTMLVersion(targetURL, content string) string {
	log.Printf("Starting detect html version for base URL : %s\n", targetURL)
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
	log.Printf("Starting extract page title for base URL: %s\n", targetURL)
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
	log.Printf("Starting count headers for base URL: %s\n", targetURL)
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
	log.Printf("Starting link analysis for base URL: %s\n", baseURL)

	//TODO validate baseURL

	tokenizer := html.NewTokenizer(strings.NewReader(content))
	// To store "href" values
	links := []string{}

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

	log.Printf("Finished HTML parsing. Found %d potential links.\n", len(links))

	result := LinkAnalysis{
		Total: len(links),
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		log.Fatalf("Error pasing provided baseURL %q: %v\n", baseURL, err)
	}

	internalCount := 0
	externalCount := 0
	inaccessibleCount := 0
	inaccessibleURLs := []string{}

	client := &http.Client{Timeout: 3 * time.Second}

	for i, link := range links {
		log.Printf("Processing link %d/%d:%s\n", i+1, len(links), link)

		u, err := url.Parse(link)
		if err != nil {
			inaccessibleCount++
			inaccessibleURLs = append(inaccessibleURLs, link)
			continue
		}

		if !u.IsAbs() {
			log.Printf("Info: Resoloving relative URL %q against base. \n", link)
			u = base.ResolveReference(u)
		}

		absoluteURL := u.String()

		if u.Host == base.Host {
			internalCount++
			log.Printf("Info: Identified INTERNAL link : %s\n", absoluteURL)
		} else {
			externalCount++
			log.Printf("Info: Identified External link : %s\n", absoluteURL)
		}

		resp, err := client.Head(absoluteURL)
		if err != nil || resp.StatusCode >= 400 {
			// Log HTTP status errors specifically
			log.Printf("Error: Recevied inaccessible status code %d for URL:%s\n", resp.StatusCode, absoluteURL)
			inaccessibleCount++
			inaccessibleURLs = append(inaccessibleURLs, absoluteURL)
		} else {
			log.Printf("Info: Link %s is accessible (Status: %d).\n", absoluteURL, resp.StatusCode)
		}
		if resp != nil {
			resp.Body.Close()
		}

	}

	result.Internal = internalCount
	result.External = externalCount
	result.Inaccessible = inaccessibleCount
	result.InaccessibleURLs = inaccessibleURLs

	log.Printf("Analysis complete. Total: %d, Internal: %d, External: %d, Inaccessible: %d.\n",
		result.Total, result.Internal, result.External, result.Inaccessible)

	return result

}

func CheckForLoginForm(content *html.Node) bool {
	var hasLogin bool

	var traverse func(*html.Node)

	traverse = func(currentNode *html.Node) {
		if hasLogin {
			return
		}

		if currentNode.Type == html.ElementNode && currentNode.Data == "form" {
			if containPassword(currentNode) {
				hasLogin = true
				return
			}
		}
		for childNode := currentNode.FirstChild; childNode != nil && !hasLogin; childNode = childNode.NextSibling {
			//traverse(currentNode)
			traverse(childNode)
			if hasLogin {
				break
			}
		}
	}
	traverse(content)
	return hasLogin
}

func containPassword(form *html.Node) bool {
	var passwordFieldFound bool

	var traverse func(*html.Node)

	traverse = func(currentNode *html.Node) {
		if passwordFieldFound {
			return
		}
		if currentNode.Type == html.ElementNode && currentNode.Data == "input" {
			// Itrate through all attribute of the input tag.
			for _, inputAttribute := range currentNode.Attr {
				// Check if the attribute is type= "password"
				if inputAttribute.Key == "type" && inputAttribute.Val == "password" {
					passwordFieldFound = true
					return
				}
			}
		}
		for childNode := currentNode.FirstChild; childNode != nil && !passwordFieldFound; childNode = childNode.NextSibling {
			//traverse(currentNode)
			traverse(childNode)
			if passwordFieldFound {
				break
			}
		}
	}
	traverse(form)
	return passwordFieldFound
}
