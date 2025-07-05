package http

import (
	"fmt"
	"net/http"
	"time"

	clihttp "web-pages-analyzer/internal/domain/clients/http"
)

type httpClient struct {
	httpClient *http.Client
}

func New(cfg *clihttp.HttpClientCfg) clihttp.HttpClient {
	return &httpClient{
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= cfg.MaxRedirects {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
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

func (c *httpClient) Head(url string) (*http.Response, error) {
	resp, err := c.httpClient.Head(url)
	if err != nil {
		return nil, clihttp.NewHttpError(
			http.StatusBadGateway,
			fmt.Sprintf("error in HEAD call: %s", err.Error()),
		)
	}

	if !isHeadSucceed(resp.StatusCode) {
		return nil, clihttp.NewHttpError(
			resp.StatusCode,
			fmt.Sprintf("faliure in HEAD call: %s", resp.Status),
		)
	}

	return resp, nil
}

func isSucceed(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

func isHeadSucceed(statusCode int) bool {
	return statusCode >= 200 && statusCode < 400
}
