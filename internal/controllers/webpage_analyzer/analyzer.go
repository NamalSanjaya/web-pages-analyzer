package webpage_analyzer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	dmpg "web-pages-analyzer/internal/domain/webpage"
)

type analyzeRequest struct {
	URL string `json:"url"`
}

type webPageAnalyzerCtrler struct {
	analyzer dmpg.WebPageAnalyzer
}

func New(wpa dmpg.WebPageAnalyzer) *webPageAnalyzerCtrler {
	return &webPageAnalyzerCtrler{analyzer: wpa}
}

func (wpac *webPageAnalyzerCtrler) Analyze(w http.ResponseWriter, r *http.Request) {
	var req analyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	if err := validateURL(req.URL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := wpac.analyzer.Analyze(req.URL)
	if err != nil {
		http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL is required")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %s", err.Error())
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("only HTTP and HTTPS are supported")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	return nil
}
