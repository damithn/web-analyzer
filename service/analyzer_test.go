package service

import (
	"testing"
)

const sampleHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Test Page</title>
</head>
<body>
    <h1>Main Heading</h1>
    <h2>Sub Heading</h2>
    <a href="/">Home</a>
    <a href="http://example.com">External</a>
    <a>Broken</a>
    <form>
        <input type="text" name="user"/>
        <input type="password" name="pass"/>
    </form>
</body>
</html>
`

func TestDetectHTMLVersion(t *testing.T) {
	version := DetectHTMLVersion("www.example.com", sampleHTML)
	if version != "HTML5" {
		t.Errorf("Expected HTML5, got %s", version)
	}
}

func TestExtractPageTitle(t *testing.T) {
	title := ExtractPageTitle("www.example.com", sampleHTML)
	if title != "Test Page" {
		t.Errorf("Expected 'Test Page', got '%s'", title)
	}
}

func TestCountHeadings(t *testing.T) {
	headings := CountHeadings("www.example.com", sampleHTML)
	if headings["h1"] != 1 || headings["h2"] != 1 {
		t.Errorf("Unexpected heading counts: %v", headings)
	}
}

func TestAnalyzeLinks(t *testing.T) {
	result := AnalyzeLinks("http://example.com", sampleHTML)
	if result.Internal != 2 || result.External != 0 || result.Inaccessible != 0 {
		t.Errorf("Expected (2,0,0), got (%d,%d,%d)", result.Internal, result.External, result.Inaccessible)
	}
}

func TestCheckForLoginForm(t *testing.T) {
	found := CheckForLoginForm(sampleHTML)
	if !found {
		t.Errorf("Expected login form to be detected")
	}
}
