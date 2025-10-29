package service

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

func AnalyzeHTMLVersion(targetURL string) (string, error) {
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = "https://" + targetURL
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(targetURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	content := strings.ToLower(string(bodyBytes))
	docTypeIndex := strings.Index(content, "<!doctype")
	if docTypeIndex == -1 {
		return "", errors.New("doctype not found; unable to determine HTML version")
	}

	switch {
	case strings.Contains(content, "<!doctype html>"):
		return "HTML5", nil
	case strings.Contains(content, "xhtml 1.0"):
		return "XHTML 1.0", nil
	case strings.Contains(content, "xhtml 1.1"):
		return "XHTML 1.1", nil
	case strings.Contains(content, "html 4.01"):
		return "HTML 4.01", nil
	case strings.Contains(content, "html 3.2"):
		return "HTML 3.2", nil
	case strings.Contains(content, "html 2.0"):
		return "HTML 2.0", nil
	default:
		return "Unknown or custom HTML version", nil
	}
}
