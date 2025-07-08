package html_parser

import (
	"bufio"
	"bytes"
	"io"
	"net/url"
	"strings"
	"sync"

	"golang.org/x/net/html"

	clihttp "web-pages-analyzer/internal/domain/clients/http"
	dmhtml "web-pages-analyzer/internal/domain/html"
	utlstr "web-pages-analyzer/internal/utils/string"
)

const (
	html5Version   = "HTML5"
	html4Version   = "HTML 4.01"
	xhtmlVersion   = "XHTML"
	unknownVersion = "Unknown HTML Version"
)

type parser struct {
	node    *html.Node
	baseUrl *url.URL
	body    io.Reader
	client  clihttp.HttpClient
}

func New(body io.Reader, baseUrl string, client clihttp.HttpClient) (dmhtml.HtmlParser, error) {
	var buf bytes.Buffer
	teeBody := io.TeeReader(body, &buf)

	node, err := html.Parse(teeBody)
	if err != nil {
		return nil, err
	}

	base, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	return &parser{
		node:    node,
		baseUrl: base,
		body:    io.NopCloser(bytes.NewReader(buf.Bytes())),
		client:  client,
	}, nil
}

func (p *parser) GetHtmlVersion() string {
	scanner := bufio.NewScanner(p.body)

	for scanner.Scan() {
		line := strings.ToLower(scanner.Text())

		if strings.Contains(line, "<!doctype html>") {
			return html5Version
		} else if strings.Contains(line, "<!doctype html public") {
			if strings.Contains(line, "4.01") {
				return html4Version
			} else if strings.Contains(line, "xhtml") {
				return xhtmlVersion
			}
		}
	}

	return unknownVersion
}

// Extract the title from the HTML document
func (p *parser) GetTitle() string {
	return strings.TrimSpace(findTitle(p.node))
}

// Count the number of heading levels in the HTML document
func (p *parser) CountHeadingLevels() map[string]int {
	headings := map[string]int{
		"h1": 0,
		"h2": 0,
		"h3": 0,
		"h4": 0,
		"h5": 0,
		"h6": 0,
	}

	countHeadingLevels(p.node, headings)
	return headings
}

func (p *parser) HasLoginForm() bool {
	return existLoginForm(p.node)
}

func (p *parser) AnalyzeLinks() *dmhtml.LinkAnalysis {
	var internal, external, inaccessible int

	links := extractLinks(p.node, p.baseUrl)

	type linkResult struct {
		isInternal   bool
		isAccessible bool
	}

	results := make(chan linkResult, len(links))
	var wg sync.WaitGroup

	for _, link := range links {
		wg.Add(1)
		go func(linkURL string, baseHost string) {
			defer wg.Done()

			parsedLink, err := url.Parse(linkURL)
			if err != nil {
				results <- linkResult{isInternal: false, isAccessible: false}
				return
			}

			isInternal := isInternalLink(parsedLink, baseHost)

			isAccessible := p.checkLinkAccessibility(linkURL)

			results <- linkResult{isInternal: isInternal, isAccessible: isAccessible}
		}(link, p.baseUrl.Host)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect data from results channel
	for result := range results {
		if result.isInternal {
			internal++
		} else {
			external++
		}

		if !result.isAccessible {
			inaccessible++
		}
	}

	return &dmhtml.LinkAnalysis{
		Internal:     internal,
		External:     external,
		Inaccessible: inaccessible,
	}
}

// check if a link is accessible
func (p *parser) checkLinkAccessibility(linkURL string) bool {
	resp, err := p.client.Head(linkURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func extractLinks(node *html.Node, base *url.URL) []string {
	var links []string

	if node.Type == html.ElementNode && node.Data == "a" {
		for _, attr := range node.Attr {
			if attr.Key == "href" && attr.Val != "" {
				if resolvedURL := resolveURL(attr.Val, base); resolvedURL != "" {
					links = append(links, resolvedURL)
				}
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		links = append(links, extractLinks(child, base)...)
	}

	return links
}

func resolveURL(href string, base *url.URL) string {
	if href == "" || utlstr.ContainsAnyPrefix(href, "#", "javascript:", "mailto:", "tel:") {
		return ""
	}

	parsed, err := url.Parse(href)
	if err != nil {
		return ""
	}

	resolved := base.ResolveReference(parsed)
	return resolved.String()
}

func isInternalLink(linkURL *url.URL, baseHost string) bool {
	if linkURL.Host == "" {
		return true
	}

	return strings.EqualFold(linkURL.Host, baseHost)
}

func existLoginForm(node *html.Node) bool {
	if node.Type == html.ElementNode {
		if node.Data == "form" && existLoginFormTag(node) {
			return true
		}

		if node.Data == "input" && existInputTag(node) {
			return true
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if existLoginForm(child) {
			return true
		}
	}

	return false
}

func existLoginFormTag(formNode *html.Node) bool {
	hasPasswordField := false
	hasUsernameField := false

	for _, attr := range formNode.Attr {
		value := strings.ToLower(attr.Val)

		if utlstr.ContainsAnySubstring(attr.Key, "action", "id", "class", "name") &&
			utlstr.ContainsAnySubstring(value, "login", "signin", "auth", "session") {
			return true
		}
	}

	checkFormInputs(formNode, &hasPasswordField, &hasUsernameField)

	return hasPasswordField && hasUsernameField
}

func checkFormInputs(node *html.Node, hasPassword *bool, hasUsername *bool) {
	if node.Type == html.ElementNode && node.Data == "input" {
		inputType := ""
		inputName := ""
		inputId := ""

		for _, attr := range node.Attr {
			switch attr.Key {
			case "type":
				inputType = strings.ToLower(attr.Val)
			case "name":
				inputName = strings.ToLower(attr.Val)
			case "id":
				inputId = strings.ToLower(attr.Val)
			}
		}

		if inputType == "password" {
			*hasPassword = true
		}

		if utlstr.ContainsAnySubstring(inputType, "email", "text") {
			if utlstr.ContainsAnySubstring(inputName, "user", "email", "login") ||
				utlstr.ContainsAnySubstring(inputId, "user", "email", "login") {
				*hasUsername = true
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		checkFormInputs(child, hasPassword, hasUsername)
	}
}

func existInputTag(inputNode *html.Node) bool {
	for _, attr := range inputNode.Attr {
		value := strings.ToLower(attr.Val)

		if attr.Key == "type" && value == "password" {
			return true
		}

		if utlstr.ContainsAnySubstring(attr.Key, "name", "id", "class") &&
			utlstr.ContainsAnySubstring(value, "password", "login", "signin", "auth") {
			return true
		}
	}

	return false
}

func countHeadingLevels(node *html.Node, headings map[string]int) {
	if node.Type == html.ElementNode {
		switch node.Data {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			headings[node.Data]++
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		countHeadingLevels(child, headings)
	}
}

// Search for the title element in the HTML document
func findTitle(node *html.Node) string {
	if node.Type == html.ElementNode && node.Data == "title" {
		return getTextContent(node)
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if title := findTitle(child); title != "" {
			return title
		}
	}

	return ""
}

func getTextContent(node *html.Node) string {
	if node.Type == html.TextNode {
		return node.Data
	}

	var text strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		text.WriteString(getTextContent(child))
	}

	return text.String()
}
