package html

import (
	"io"

	clihttp "web-pages-analyzer/internal/domain/clients/http"
)

type ParserFactory interface {
	CreateParser(body io.Reader, baseUrl string, client clihttp.HttpClient) (HtmlParser, error)
}
