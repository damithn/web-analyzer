package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"web-analyzer/service"
)

func TestAnalyzeWebPageIntegration(t *testing.T) {
	// Setup a mock HTML page (acts as a fake website)
	mockHTML := `
	<!DOCTYPE html>
	<html>
	<head><title>Integration Test</title></head>
	<body>
		<h1>Main</h1>
		<h2>Sub</h2>
		<a href="/">Home</a>
		<a href="http://external.com">External</a>
		<form>
			<input type="text" name="user"/>
			<input type="password" name="pass"/>
		</form>
	</body>
	</html>
	`

	// Create a mock server (no external network calls)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockHTML))
	}))
	defer server.Close()

	// Call the actual service under test
	result, err := service.AnalyzeWebPage(server.URL)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Validate the complete flow
	if result.HTMLVersion != "HTML5" {
		t.Errorf("Expected HTML5, got %s", result.HTMLVersion)
	}
	if result.PageTitle != "Integration Test" {
		t.Errorf("Expected title 'Integration Test', got %s", result.PageTitle)
	}
	if result.Headings["h1"] != 1 || result.Headings["h2"] != 1 {
		t.Errorf("Unexpected headings: %v", result.Headings)
	}
	if result.Links.Internal != 0 || result.Links.External != 1 {
		t.Errorf("Unexpected link counts: Internal=%d, External=%d", result.Links.Internal, result.Links.External)
	}
	if !result.ContainsLoginForm {
		t.Errorf("Expected login form to be detected")
	}
}
