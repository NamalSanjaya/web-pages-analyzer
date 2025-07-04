package html_parser

import (
	"bufio"
	"io"
	"strings"

	"golang.org/x/net/html"
)

const (
	html5Version   = "HTML5"
	html4Version   = "HTML 4.01"
	xhtmlVersion   = "XHTML"
	unknownVersion = "Unknown HTML Version"
)

type parser struct {
	node *html.Node
	body io.Reader
}

func New(body io.Reader) (*parser, error) {
	node, err := html.Parse(body)
	if err != nil {
		return nil, err
	}

	return &parser{
		node: node,
		body: body,
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

func (p *parser) GetTitle() string {
	return strings.TrimSpace(findTitle(p.node))
}

// Search for the title element in the HTML document
func findTitle(node *html.Node) string {
	if node.Type == html.ElementNode && node.Data == "title" {
		return getTextContent(node)
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if title := findTitle(c); title != "" {
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
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		text.WriteString(getTextContent(c))
	}

	return text.String()
}
