package http

import (
	"net/http"
)

type HttpClientCfg struct {
	Timeout      int // Timeout in seconds
	MaxRedirects int
	Transport    http.RoundTripper
}

type HttpClient interface {
	Get(url string) (*http.Response, error)
	Head(url string) (*http.Response, error)
}
