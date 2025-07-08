package html_parser

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	clihttp "web-pages-analyzer/internal/domain/clients/http"
	"web-pages-analyzer/internal/infrastructure/clients/http/mocks"
	httpmocks "web-pages-analyzer/internal/infrastructure/clients/http/mocks"
)

func Test_GetHtmlVersion(t *testing.T) {
	tests := []struct {
		name        string
		htmlContent string
		expected    string
	}{
		{
			name:        "HTML5 HTML version",
			htmlContent: "<!DOCTYPE html>\n<html><head><title>Test</title></head></html>",
			expected:    "HTML5",
		},
		{
			name:        "HTML5 HTML version - case insensitive",
			htmlContent: "<!doctype HTML>\n<html><head><title>Test</title></head></html>",
			expected:    "HTML5",
		},
		{
			name:        "HTML 4.01 HTML version",
			htmlContent: `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd">`,
			expected:    "HTML 4.01",
		},
		{
			name:        "XHTML HTML version",
			htmlContent: `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">`,
			expected:    "XHTML",
		},
		{
			name:        "Unknown HTML version",
			htmlContent: "<html><head><title>Test</title></head></html>",
			expected:    "Unknown HTML Version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := httpmocks.NewMockHttpClient(ctrl)
			body := strings.NewReader(tt.htmlContent)

			parser, err := New(body, "https://example.com", mockClient)
			if err != nil {
				t.Fatalf("failed to create new parser: %v", err)
			}

			result := parser.GetHtmlVersion()
			if result != tt.expected {
				t.Errorf("expected HTML version %s, got %s", tt.expected, result)
			}
		})
	}
}

func Test_GetTitle(t *testing.T) {
	tests := []struct {
		name        string
		htmlContent string
		expected    string
	}{
		{
			name:        "valid title",
			htmlContent: "<html><head><title>Test Page</title></head></html>",
			expected:    "Test Page",
		},
		{
			name:        "title with whitespace",
			htmlContent: "<html><head><title>  Test Page  </title></head></html>",
			expected:    "Test Page",
		},
		{
			name:        "no title",
			htmlContent: "<html><head></head></html>",
			expected:    "",
		},
		{
			name:        "empty title",
			htmlContent: "<html><head><title></title></head></html>",
			expected:    "",
		},
		{
			name:        "multiple titles",
			htmlContent: "<html><head><title>First Title</title><title>Second Title</title></head></html>",
			expected:    "First Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := httpmocks.NewMockHttpClient(ctrl)
			body := strings.NewReader(tt.htmlContent)

			parser, err := New(body, "https://example.com", mockClient)
			if err != nil {
				t.Fatalf("failed to create new parser: %v", err)
			}

			result := parser.GetTitle()
			if result != tt.expected {
				t.Errorf("expected title %s, got %q", tt.expected, result)
			}
		})
	}
}

func Test_CountHeadingLevels(t *testing.T) {
	tests := []struct {
		name        string
		htmlContent string
		expected    map[string]int
	}{
		{
			name: "various headings",
			htmlContent: `<html><body>
				<h1>Heading 1</h1>
				<h2>Heading 2</h2>
				<h2>Another H2</h2>
				<h3>Heading 3</h3>
				<h6>Heading 6</h6>
			</body></html>`,
			expected: map[string]int{
				"h1": 1, "h2": 2, "h3": 1, "h4": 0, "h5": 0, "h6": 1,
			},
		},
		{
			name:        "no headings",
			htmlContent: "<html><body><p>No headings here</p></body></html>",
			expected: map[string]int{
				"h1": 0, "h2": 0, "h3": 0, "h4": 0, "h5": 0, "h6": 0,
			},
		},
		{
			name: "nested headings",
			htmlContent: `<html><body>
				<div><h1>Main</h1><div><h2>Sub</h2></div></div>
			</body></html>`,
			expected: map[string]int{
				"h1": 1, "h2": 1, "h3": 0, "h4": 0, "h5": 0, "h6": 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockHttpClient(ctrl)
			body := strings.NewReader(tt.htmlContent)

			parser, err := New(body, "https://example.com", mockClient)
			if err != nil {
				t.Fatalf("failed to create new parser: %v", err)
			}

			result := parser.CountHeadingLevels()
			for level, expectedCount := range tt.expected {
				if result[level] != expectedCount {
					t.Errorf("expected %s count %d, got %d", level, expectedCount, result[level])
				}
			}
		})
	}
}

func Test_HasLoginForm(t *testing.T) {
	tests := []struct {
		name        string
		htmlContent string
		expected    bool
	}{
		{
			name: "with login form",
			htmlContent: `<html><body>
				<form action="/login" method="post">
					<input type="text" name="username">
					<input type="password" name="password">
				</form>
			</body></html>`,
			expected: true,
		},
		{
			name: "login form with signin class",
			htmlContent: `<html><body>
				<form class="signin-form" method="post">
					<input type="email" name="email">
					<input type="password" name="password">
				</form>
			</body></html>`,
			expected: true,
		},
		{
			name: "login form with password input",
			htmlContent: `<html><body>
				<input type="password" name="login-password">
			</body></html>`,
			expected: true,
		},
		{
			name: "form without login",
			htmlContent: `<html><body>
				<form method="post">
					<input type="text" name="name">
					<input type="email" name="email">
				</form>
			</body></html>`,
			expected: false,
		},
		{
			name:        "no forms",
			htmlContent: `<html><body><p>No forms here</p></body></html>`,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := httpmocks.NewMockHttpClient(ctrl)
			body := strings.NewReader(tt.htmlContent)

			parser, err := New(body, "https://example.com", mockClient)
			if err != nil {
				t.Fatalf("unexpected error creating parser: %v", err)
			}

			result := parser.HasLoginForm()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func Test_AnalyzeLinks(t *testing.T) {
	tests := []struct {
		name                 string
		htmlContent          string
		baseURL              string
		mockSetup            func(*mocks.MockHttpClient)
		expectedInternal     int
		expectedExternal     int
		expectedInaccessible int
	}{
		{
			name: "mixed internal and external links",
			htmlContent: `<html><body>
				<a href="/internal">Internal Link</a>
				<a href="https://example.com/page">Same Domain</a>
				<a href="https://external.com">External Link</a>
				<a href="mailto:test@example.com">Email</a>
				<a href="#anchor">Anchor</a>
			</body></html>`,
			baseURL: "https://example.com",
			mockSetup: func(mock *mocks.MockHttpClient) {
				// Internal links
				mock.EXPECT().Head("https://example.com/internal").Return(
					&http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(""))}, nil)
				mock.EXPECT().Head("https://example.com/page").Return(
					&http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(""))}, nil)
				// External link
				mock.EXPECT().Head("https://external.com").Return(
					&http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(""))}, nil)
			},
			expectedInternal:     2,
			expectedExternal:     1,
			expectedInaccessible: 1,
		},
		{
			name: "all accessible links",
			htmlContent: `<html><body>
				<a href="/page1">Page 1</a>
				<a href="https://external.com">External</a>
			</body></html>`,
			baseURL: "https://example.com",
			mockSetup: func(mock *mocks.MockHttpClient) {
				mock.EXPECT().Head("https://example.com/page1").Return(
					&http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(""))}, nil)
				mock.EXPECT().Head("https://external.com").Return(
					&http.Response{StatusCode: 301, Body: io.NopCloser(strings.NewReader(""))}, nil)
			},
			expectedInternal:     1,
			expectedExternal:     1,
			expectedInaccessible: 0,
		},
		{
			name: "network errors make links inaccessible",
			htmlContent: `<html><body>
				<a href="/page1">Page 1</a>
				<a href="https://unreachable.com">Unreachable</a>
			</body></html>`,
			baseURL: "https://example.com",
			mockSetup: func(mock *mocks.MockHttpClient) {
				mock.EXPECT().Head("https://example.com/page1").Return(
					&http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(""))}, nil)
				mock.EXPECT().Head("https://unreachable.com").Return(
					nil, clihttp.NewHttpError(502, "Bad Gateway"))
			},
			expectedInternal:     1,
			expectedExternal:     1,
			expectedInaccessible: 1,
		},
		{
			name:        "no links",
			htmlContent: `<html><body><p>No links here</p></body></html>`,
			baseURL:     "https://example.com",
			mockSetup: func(mock *mocks.MockHttpClient) {
				// No expectations
			},
			expectedInternal:     0,
			expectedExternal:     0,
			expectedInaccessible: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := httpmocks.NewMockHttpClient(ctrl)
			tt.mockSetup(mockClient)

			body := strings.NewReader(tt.htmlContent)
			parser, err := New(body, tt.baseURL, mockClient)
			if err != nil {
				t.Fatalf("unexpected error creating parser: %v", err)
			}

			result := parser.AnalyzeLinks()

			if result.Internal != tt.expectedInternal {
				t.Errorf("expected %d internal links, got %d", tt.expectedInternal, result.Internal)
			}
			if result.External != tt.expectedExternal {
				t.Errorf("expected %d external links, got %d", tt.expectedExternal, result.External)
			}
			if result.Inaccessible != tt.expectedInaccessible {
				t.Errorf("expected %d inaccessible links, got %d", tt.expectedInaccessible, result.Inaccessible)
			}
		})
	}
}
