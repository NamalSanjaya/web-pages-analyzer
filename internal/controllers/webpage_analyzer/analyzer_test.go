package webpage_analyzer

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	dmhtml "web-pages-analyzer/internal/domain/html"
	dmpg "web-pages-analyzer/internal/domain/webpage"
	mocks "web-pages-analyzer/internal/usecases/webpage_analyzer/mocks"
)

func equalHeadings(a, b map[string]int) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func Test_Analyze_Success(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		url            string
		analysisResult *dmpg.WebPageAnalysis
	}{
		{
			name:        "HTML5 page-1 Analysis",
			requestBody: `{"url": "https://example.com"}`,
			url:         "https://example.com",
			analysisResult: &dmpg.WebPageAnalysis{
				HTMLVersion:  "HTML5",
				Title:        "Title-1",
				Headings:     map[string]int{"h1": 2, "h2": 3, "h3": 1, "h4": 0, "h5": 0, "h6": 0},
				Links:        dmhtml.LinkAnalysis{Internal: 5, External: 3, Inaccessible: 1},
				HasLoginForm: true,
			},
		},
		{
			name:        "HTML5 page-2 Analysis",
			requestBody: `{"url": "https://example2.com"}`,
			url:         "https://example2.com",
			analysisResult: &dmpg.WebPageAnalysis{
				HTMLVersion:  "HTML 4.01",
				Title:        "Title-2",
				Headings:     map[string]int{"h1": 1, "h2": 0, "h3": 0, "h4": 0, "h5": 0, "h6": 0},
				Links:        dmhtml.LinkAnalysis{Internal: 0, External: 0, Inaccessible: 0},
				HasLoginForm: false,
			},
		},
		{
			name:        "HTML5 page-3 Analysis",
			requestBody: `{"url": "https://example3.com"}`,
			url:         "https://example3.com",
			analysisResult: &dmpg.WebPageAnalysis{
				HTMLVersion:  "XHTML",
				Title:        "Title-3",
				Headings:     map[string]int{"h1": 1, "h2": 2, "h3": 3, "h4": 1, "h5": 0, "h6": 1},
				Links:        dmhtml.LinkAnalysis{Internal: 10, External: 5, Inaccessible: 3},
				HasLoginForm: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAnalyzer := mocks.NewMockWebPageAnalyzer(ctrl)
			mockAnalyzer.EXPECT().
				Analyze(tt.url).
				Return(tt.analysisResult, nil).
				Times(1)

			controller := New(mockAnalyzer)

			req := httptest.NewRequest(http.MethodPost, "/api/analyze", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			controller.Analyze(w, req)

			// Verify results
			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			var result dmpg.WebPageAnalysis
			if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if result.HTMLVersion != tt.analysisResult.HTMLVersion {
				t.Errorf("expected HTML version %q, got %q", tt.analysisResult.HTMLVersion, result.HTMLVersion)
			}

			if result.Title != tt.analysisResult.Title {
				t.Errorf("expected title %q, got %q", tt.analysisResult.Title, result.Title)
			}

			if !equalHeadings(result.Headings, tt.analysisResult.Headings) {
				t.Errorf("expected headings %v, got %v", tt.analysisResult.Headings, result.Headings)
			}

			if result.Links.Internal != tt.analysisResult.Links.Internal {
				t.Errorf("expected %d internal links, got %d", tt.analysisResult.Links.Internal, result.Links.Internal)
			}

			if result.Links.External != tt.analysisResult.Links.External {
				t.Errorf("expected %d external links, got %d", tt.analysisResult.Links.External, result.Links.External)
			}

			if result.Links.Inaccessible != tt.analysisResult.Links.Inaccessible {
				t.Errorf("expected %d inaccessible links, got %d", tt.analysisResult.Links.Inaccessible, result.Links.Inaccessible)
			}

			if result.HasLoginForm != tt.analysisResult.HasLoginForm {
				t.Errorf("expected a login form %v, got %v", tt.analysisResult.HasLoginForm, result.HasLoginForm)
			}
		})
	}
}

func Test_Analyze_ErrDecodingReqBody(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
	}{
		{
			name:        "malformed JSON",
			requestBody: `{"url": "https://example.com"`,
		},
		{
			name:        "non-JSON content",
			requestBody: `this is not json`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAnalyzer := mocks.NewMockWebPageAnalyzer(ctrl)

			controller := New(mockAnalyzer)

			req := httptest.NewRequest(http.MethodPost, "/api/analyze", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			controller.Analyze(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func Test_Analyze_ErrURLValidation(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectedError string
	}{
		{
			name:          "empty URL",
			url:           "",
			expectedError: "URL is required",
		},
		{
			name:          "invalid URL format",
			url:           "example.com",
			expectedError: "only HTTP and HTTPS are supported",
		},
		{
			name:          "missing host",
			url:           "https://",
			expectedError: "host cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAnalyzer := mocks.NewMockWebPageAnalyzer(ctrl)

			controller := New(mockAnalyzer)

			requestBody := map[string]string{"url": tt.url}
			jsonBody, _ := json.Marshal(requestBody)

			req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			controller.Analyze(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
			}

			responseBody := strings.TrimSpace(w.Body.String())
			if !strings.Contains(responseBody, tt.expectedError) {
				t.Errorf("expected error message should contain %q, got %q", tt.expectedError, responseBody)
			}
		})
	}
}

func TestAnalyze_AnalyzerError(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		analyzerError error
		expectedError string
	}{
		{
			name:          "parsing error",
			url:           "https://example.com",
			analyzerError: errors.New("failed to parse HTML"),
			expectedError: "Internal server error: failed to parse HTML",
		},
		{
			name:          "unknown error",
			url:           "https://example.com",
			analyzerError: errors.New("something went wrong"),
			expectedError: "Internal server error: something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAnalyzer := mocks.NewMockWebPageAnalyzer(ctrl)
			mockAnalyzer.EXPECT().
				Analyze(tt.url).
				Return(nil, tt.analyzerError).
				Times(1)

			controller := New(mockAnalyzer)

			requestBody := map[string]string{"url": tt.url}
			jsonBody, _ := json.Marshal(requestBody)

			req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			controller.Analyze(w, req)

			if w.Code != http.StatusInternalServerError {
				t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
			}

			responseBody := strings.TrimSpace(w.Body.String())
			if responseBody != tt.expectedError {
				t.Errorf("expected error %q, got %q", tt.expectedError, responseBody)
			}
		})
	}
}
