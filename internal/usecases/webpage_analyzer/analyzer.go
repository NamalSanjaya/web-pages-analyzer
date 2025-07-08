package webpage_analyzer

import (
	clihttp "web-pages-analyzer/internal/domain/clients/http"
	dmhtml "web-pages-analyzer/internal/domain/html"
	dmpg "web-pages-analyzer/internal/domain/webpage"
)

type webPageAnalyzer struct {
	httpClient    clihttp.HttpClient
	parserFactory dmhtml.ParserFactory
}

func New(httpClient clihttp.HttpClient, parserFactory dmhtml.ParserFactory) dmpg.WebPageAnalyzer {
	return &webPageAnalyzer{
		httpClient:    httpClient,
		parserFactory: parserFactory,
	}
}

func (wpa *webPageAnalyzer) Analyze(url string) (*dmpg.WebPageAnalysis, error) {
	// Fetch the web page
	resp, err := wpa.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	parser, err := wpa.parserFactory.CreateParser(resp.Body, url, wpa.httpClient)
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
