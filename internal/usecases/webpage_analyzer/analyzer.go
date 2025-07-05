package webpage_analyzer

import (
	clihttp "web-pages-analyzer/internal/domain/clients/http"
	dmpg "web-pages-analyzer/internal/domain/webpage"
	htmpar "web-pages-analyzer/internal/infrastructure/html_parser"
)

type webPageAnalyzer struct {
	httpClient clihttp.HttpClient
}

func New(httpClient clihttp.HttpClient) *webPageAnalyzer {
	return &webPageAnalyzer{
		httpClient: httpClient,
	}
}

func (wpa *webPageAnalyzer) Analyze(url string) (*dmpg.WebPageAnalysis, error) {
	// Fetch the web page
	resp, err := wpa.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	parser, err := htmpar.New(resp.Body, url, wpa.httpClient)
	if err != nil {
		return nil, err
	}

	return &dmpg.WebPageAnalysis{
		HTMLVersion:  parser.GetHtmlVersion(),
		Title:        parser.GetTitle(),
		Headings:     parser.CountHeadingLevels(),
		Links:        *parser.AnalyzeLinks(),
		HasLoginForm: parser.HasLoginForm(),
	}, nil
}
