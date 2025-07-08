package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	clihttp "web-pages-analyzer/internal/domain/clients/http"
)

type mockTrasporter struct{}

func (m *mockTrasporter) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("[mockTrasporter]: failed to resolve DNS")
}

func validateResults(t *testing.T, resp *http.Response, err error, expectedError bool, statusCode int) {
	if expectedError {
		if err == nil {
			t.Error("expected error, got nil")
		}
		if resp != nil {
			t.Errorf("expected response nil, got %v", resp)
		}
	}

	if !expectedError {
		if err != nil {
			t.Errorf("expected no error, got an error: %v", err)
		}
		if resp == nil {
			t.Error("expected response non-nil, got nil")
		}
		if resp != nil && resp.StatusCode != statusCode {
			t.Errorf("expected status code %d, got %d", statusCode, resp.StatusCode)
		}
	}
}

func Test_HttpClient_Get(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectedError bool
	}{
		{
			name:          "GET request with 200",
			statusCode:    http.StatusOK,
			responseBody:  "<html><body>Test</body></html>",
			expectedError: false,
		},
		{
			name:          "GET request with 204",
			statusCode:    http.StatusNoContent,
			responseBody:  "Created",
			expectedError: false,
		},
		{
			name:          "GET request with 400",
			statusCode:    http.StatusBadRequest,
			responseBody:  "Bad Request",
			expectedError: true,
		},
		{
			name:          "GET request with 500",
			statusCode:    http.StatusInternalServerError,
			responseBody:  "Internal Server Error",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET request, got %s", r.Method)
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create client
			cfg := &clihttp.HttpClientCfg{
				Timeout:      10,
				MaxRedirects: 5,
			}
			client := New(cfg)

			resp, err := client.Get(server.URL)

			// Verify results
			validateResults(t, resp, err, tt.expectedError, tt.statusCode)
			if resp != nil {
				resp.Body.Close()
			}
		})
	}
}

func Test_HttpClient_Get_Error(t *testing.T) {
	// Create client
	cfg := &clihttp.HttpClientCfg{
		Timeout:      1,
		MaxRedirects: 5,
		Transport:    &mockTrasporter{}, // mock transport to simulate DNS resolution error
	}
	client := New(cfg)

	// Make request to invalid URL
	resp, err := client.Get("http://non-existing-url.com")

	// Verify results
	if err == nil {
		t.Error("expected error but got none")
		return
	}

	if resp != nil {
		t.Error("expected nil response: got non-nil response")
		defer resp.Body.Close()
	}

	httpErr, ok := clihttp.NewHttpErrorFromErr(err)
	if !ok {
		t.Error("expected HttpError,  got something else")
	}

	if httpErr.StatusCode != http.StatusBadGateway {
		t.Errorf("expected status code %d, got %d", http.StatusBadGateway, httpErr.StatusCode)
	}
}

func TestHttpClient_Head_Success(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		expectedError bool
	}{
		{
			name:          "HEAD request with 200",
			statusCode:    http.StatusOK,
			expectedError: false,
		},
		{
			name:          "HEAD request with 301",
			statusCode:    http.StatusMovedPermanently,
			expectedError: false,
		},
		{
			name:          "HEAD request with 404",
			statusCode:    http.StatusNotFound,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodHead {
					t.Errorf("expected HEAD request, got %s", r.Method)
				}
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			// Create client
			cfg := &clihttp.HttpClientCfg{
				Timeout:      10,
				MaxRedirects: 5,
			}
			client := New(cfg)

			resp, err := client.Head(server.URL)

			// Verify results
			validateResults(t, resp, err, tt.expectedError, tt.statusCode)
			if resp != nil {
				resp.Body.Close()
			}
		})
	}
}

func Test_HttpClient_Head_Error(t *testing.T) {
	// Create client
	cfg := &clihttp.HttpClientCfg{
		Timeout:      1,
		MaxRedirects: 5,
		Transport:    &mockTrasporter{}, // mock transport to simulate DNS resolution error
	}
	client := New(cfg)

	// Make request to invalid URL
	resp, err := client.Head("http://non-existing-url.com")

	// Verify results
	if err == nil {
		t.Error("expected error but got none")
		return
	}

	if resp != nil {
		t.Error("expected nil response: got non-nil response")
		defer resp.Body.Close()
	}

	httpErr, ok := clihttp.NewHttpErrorFromErr(err)
	if !ok {
		t.Error("expected HttpError,  got something else")
	}

	if httpErr.StatusCode != http.StatusBadGateway {
		t.Errorf("expected status code %d, got %d", http.StatusBadGateway, httpErr.StatusCode)
	}
}
