package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"web-analyzer/service"
)

// JSON request body
type AnalyzeRequest struct {
	URL string `json:"url"`
}

// JSON response payload
type AnalyzeResponse struct {
	HTMLVersion string               `json:"htmlVersion"`
	PageTitle   string               `json:"pageTitle"`
	Headings    map[string]int       `json:"headings"`
	Links       service.LinkAnalysis `json:"links"`
	Error       string               `json:"error,omitempty"`
}

// JSON response with given status code
func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// Validates and normalizes a user-input URL
func validateAndNormalizeURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", errors.New("missing URL — please provide a valid URL like https://example.com")
	}

	// Prepend https:// if scheme missing
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || parsed.Host == "" {
		return "", errors.New("invalid URL format — please enter a valid URL like https://example.com")
	}

	return rawURL, nil
}

// handle / analyze POST request
func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, AnalyzeResponse{Error: "method not allowed"})
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, AnalyzeResponse{Error: "invalid JSON payload"})
		return
	}

	normalizedURL, err := validateAndNormalizeURL(req.URL)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, AnalyzeResponse{Error: err.Error()})
		return
	}

	result, err := service.AnalyzeWebPage(normalizedURL)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, AnalyzeResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, AnalyzeResponse{
		HTMLVersion: result.HTMLVersion,
		PageTitle:   result.PageTitle,
		Headings:    result.Headings,
		Links:       result.Links,
	})
}
