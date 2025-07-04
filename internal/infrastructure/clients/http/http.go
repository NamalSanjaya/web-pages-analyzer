package http

import (
	"fmt"
	"net/http"

	clihttp "web-pages-analyzer/internal/domain/clients/http"
)

type httpClient struct {
	httpClient *http.Client
}

func New() clihttp.HttpClient {
	return &httpClient{
		httpClient: http.DefaultClient,
	}
}

func (c *httpClient) Get(url string) (*http.Response, error) {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, clihttp.NewHttpError(
			http.StatusBadGateway,
			fmt.Sprintf("error in GET call: %s", err.Error()),
		)
	}

	if !isSucceed(resp.StatusCode) {
		return nil, clihttp.NewHttpError(
			resp.StatusCode,
			fmt.Sprintf("faliure in GET call: %s", resp.Status),
		)
	}

	return resp, nil
}

func isSucceed(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}
