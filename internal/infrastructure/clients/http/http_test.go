package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	clihttp "web-pages-analyzer/internal/domain/clients/http"
)

func Test_HttpClient_Get(t *testing.T) {
	tests := []struct {
		name               string
		statusCode         int
		responseBody       string
		expectedError      bool
		expectedStatusCode int
	}{
		{
			name:               "GET request with 200",
			statusCode:         http.StatusOK,
			responseBody:       "<html><body>Test</body></html>",
			expectedError:      false,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "GET request with 204",
			statusCode:         http.StatusNoContent,
			responseBody:       "Created",
			expectedError:      false,
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "GET request with 400",
			statusCode:         http.StatusBadRequest,
			responseBody:       "Bad Request",
			expectedError:      true,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "GET request with 500",
			statusCode:         http.StatusInternalServerError,
			responseBody:       "Internal Server Error",
			expectedError:      true,
			expectedStatusCode: http.StatusInternalServerError,
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
			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if resp != nil {
					t.Errorf("expected response nil, got %v", resp)
				}
			}

			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got an error: %v", err)
				}
				if resp == nil {
					t.Error("expected response non-nil, got nil")
				}
				if resp != nil && resp.StatusCode != tt.statusCode {
					t.Errorf("expected status code %d, got %d", tt.statusCode, resp.StatusCode)
				}
			}

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
