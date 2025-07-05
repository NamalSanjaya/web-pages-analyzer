package webpage

import (
	dmhtml "web-pages-analyzer/internal/domain/html"
)

type WebPageAnalysis struct {
	HTMLVersion  string              `json:"html_version"`
	Title        string              `json:"title"`
	Headings     map[string]int      `json:"headings"`
	Links        dmhtml.LinkAnalysis `json:"links"`
	HasLoginForm bool                `json:"has_login_form"`
}
