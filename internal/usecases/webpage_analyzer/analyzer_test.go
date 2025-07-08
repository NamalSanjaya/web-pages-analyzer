package webpage_analyzer

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	clihttp "web-pages-analyzer/internal/domain/clients/http"
	dmhtml "web-pages-analyzer/internal/domain/html"
	httpmocks "web-pages-analyzer/internal/infrastructure/clients/http/mocks"
	htmlmocks "web-pages-analyzer/internal/infrastructure/html_parser/mocks"
)

// Helper methods
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
		name                string
		url                 string
		responseBody        string
		expectedHTMLVersion string
		expectedTitle       string
		expectedHeadings    map[string]int
		expectedLinks       dmhtml.LinkAnalysis
		expectedLoginForm   bool
	}{
		{
			name:                "HTML page 1",
			url:                 "https://example.com",
			responseBody:        "<html><head><title>Test Page</title></head><body><h1>Header-1</h1><form action='/login'><input type='password'/></form></body></html>",
			expectedHTMLVersion: "HTML5",
			expectedTitle:       "Test Page",
			expectedHeadings:    map[string]int{"h1": 1, "h2": 0, "h3": 0, "h4": 0, "h5": 0, "h6": 0},
			expectedLinks:       dmhtml.LinkAnalysis{Internal: 2, External: 1, Inaccessible: 0},
			expectedLoginForm:   true,
		},
		{
			name:                "HTML page 2",
			url:                 "https://example.com",
			responseBody:        "<!DOCTYPE html><html><head><title>Simple</title></head><body><h2>Content</h2></body></html>",
			expectedHTMLVersion: "HTML5",
			expectedTitle:       "Simple",
			expectedHeadings:    map[string]int{"h1": 0, "h2": 1, "h3": 0, "h4": 0, "h5": 0, "h6": 0},
			expectedLinks:       dmhtml.LinkAnalysis{Internal: 0, External: 0, Inaccessible: 0},
			expectedLoginForm:   false,
		},
		{
			name:                "HTML page 3",
			url:                 "https://example.com",
			responseBody:        "<html><head><title>Blog</title></head><body><h1>Main</h1><h2>Sub1</h2><h2>Sub2</h2><h3>Detail</h3></body></html>",
			expectedHTMLVersion: "Unknown HTML Version",
			expectedTitle:       "Blog",
			expectedHeadings:    map[string]int{"h1": 1, "h2": 2, "h3": 1, "h4": 0, "h5": 0, "h6": 0},
			expectedLinks:       dmhtml.LinkAnalysis{Internal: 1, External: 2, Inaccessible: 1},
			expectedLoginForm:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHttpClient := httpmocks.NewMockHttpClient(ctrl)
			mockHttpClient.EXPECT().
				Get(tt.url).
				Return(&http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(tt.responseBody)),
				}, nil).
				Times(1)

			mockParserFactory := htmlmocks.NewMockParserFactory(ctrl)
			mockParser := htmlmocks.NewMockHtmlParser(ctrl)

			mockParserFactory.EXPECT().
				CreateParser(gomock.Any(), tt.url, mockHttpClient).
				Return(mockParser, nil).
				Times(1)

			mockParser.EXPECT().GetHtmlVersion().Return(tt.expectedHTMLVersion).Times(1)
			mockParser.EXPECT().GetTitle().Return(tt.expectedTitle).Times(1)
			mockParser.EXPECT().CountHeadingLevels().Return(tt.expectedHeadings).Times(1)
			mockParser.EXPECT().AnalyzeLinks().Return(&tt.expectedLinks).Times(1)
			mockParser.EXPECT().HasLoginForm().Return(tt.expectedLoginForm).Times(1)

			analyzer := New(mockHttpClient, mockParserFactory)
			result, err := analyzer.Analyze(tt.url)

			// Verify results
			if err != nil {
				t.Fatalf("expected nil error: got %v", err)
			}

			if result == nil {
				t.Fatal("expected result to be non-nil")
			}

			if result.HTMLVersion != tt.expectedHTMLVersion {
				t.Errorf("expected HTML version %q, got %q", tt.expectedHTMLVersion, result.HTMLVersion)
			}

			if result.Title != tt.expectedTitle {
				t.Errorf("expected title %q, got %q", tt.expectedTitle, result.Title)
			}

			if !equalHeadings(result.Headings, tt.expectedHeadings) {
				t.Errorf("expected headings %v, got %v", tt.expectedHeadings, result.Headings)
			}

			if result.Links.Internal != tt.expectedLinks.Internal {
				t.Errorf("expected %d internal links, got %d", tt.expectedLinks.Internal, result.Links.Internal)
			}

			if result.Links.External != tt.expectedLinks.External {
				t.Errorf("expected %d external links, got %d", tt.expectedLinks.External, result.Links.External)
			}

			if result.Links.Inaccessible != tt.expectedLinks.Inaccessible {
				t.Errorf("expected %d inaccessible links, got %d", tt.expectedLinks.Inaccessible, result.Links.Inaccessible)
			}

			if result.HasLoginForm != tt.expectedLoginForm {
				t.Errorf("expected login form %v, got %v", tt.expectedLoginForm, result.HasLoginForm)
			}
		})
	}
}

func Test_Analyze_HttpClientError(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		httpError     error
		expectedError string
	}{
		{
			name:          "network error",
			url:           "https://example.com",
			httpError:     clihttp.NewHttpError(502, "Bad Gateway"),
			expectedError: "Bad Gateway",
		},
		{
			name:          "404 not found error",
			url:           "https://example.com/path1/not-found",
			httpError:     clihttp.NewHttpError(404, "Not Found"),
			expectedError: "Not Found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHttpClient := httpmocks.NewMockHttpClient(ctrl)
			mockHttpClient.EXPECT().
				Get(tt.url).
				Return(nil, tt.httpError).
				Times(1)

			mockParserFactory := htmlmocks.NewMockParserFactory(ctrl)

			analyzer := New(mockHttpClient, mockParserFactory)
			result, err := analyzer.Analyze(tt.url)

			// Verify results
			if err == nil {
				t.Fatal("expected error, got nil error")
			}

			if result != nil {
				t.Fatal("expected nil result")
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("expected error to contain %q, got %q", tt.expectedError, err.Error())
			}
		})
	}
}

func TestAnalyze_ParserFactoryError(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		responseBody  string
		parserError   error
		expectedError string
	}{
		{
			name:          "invalid HTML",
			url:           "https://example.com",
			responseBody:  "invalid html body",
			parserError:   errors.New("failed to parse HTML"),
			expectedError: "failed to parse HTML",
		},
		{
			name:          "invalid URL",
			url:           "invalid-url.com",
			responseBody:  "<html><body>sample content</body></html>",
			parserError:   errors.New("invalid URL format"),
			expectedError: "invalid URL format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHttpClient := httpmocks.NewMockHttpClient(ctrl)
			mockHttpClient.EXPECT().
				Get(tt.url).
				Return(&http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(tt.responseBody)),
				}, nil).
				Times(1)

			mockParserFactory := htmlmocks.NewMockParserFactory(ctrl)
			mockParserFactory.EXPECT().
				CreateParser(gomock.Any(), tt.url, mockHttpClient).
				Return(nil, tt.parserError).
				Times(1)

			analyzer := New(mockHttpClient, mockParserFactory)
			result, err := analyzer.Analyze(tt.url)

			// Verify results
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if result != nil {
				t.Fatal("expected nil result")
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("expected error to contain %q, got %q", tt.expectedError, err.Error())
			}
		})
	}
}
