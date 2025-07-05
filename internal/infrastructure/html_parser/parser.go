package html_parser

import (
	"bufio"
	"io"
	"strings"

	"golang.org/x/net/html"

	utlstr "web-pages-analyzer/internal/utils/string"
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

func (p *parser) HasLoginForm(body io.Reader) bool {
	return hasLoginForm(p.node)
}

func hasLoginForm(node *html.Node) bool {
	if node.Type == html.ElementNode {
		if node.Data == "form" && isLoginForm(node) {
			return true
		}

		if node.Data == "input" && isLoginInput(node) {
			return true
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if hasLoginForm(child) {
			return true
		}
	}

	return false
}

func isLoginForm(formNode *html.Node) bool {
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

func isLoginInput(inputNode *html.Node) bool {
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
