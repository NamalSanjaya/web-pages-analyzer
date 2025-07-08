package html_parser

import (
	"io"

	clihttp "web-pages-analyzer/internal/domain/clients/http"
	dmhtml "web-pages-analyzer/internal/domain/html"
)

type parserFactory struct{}

func NewParserFactory() dmhtml.ParserFactory {
	return &parserFactory{}
}

func (pf *parserFactory) CreateParser(body io.Reader, baseUrl string, client clihttp.HttpClient) (dmhtml.HtmlParser, error) {
	return New(body, baseUrl, client)
}
