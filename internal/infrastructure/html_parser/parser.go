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
