package html

type LinkAnalysis struct {
	Internal     int `json:"internal"`
	External     int `json:"external"`
	Inaccessible int `json:"inaccessible"`
}

type HtmlParser interface {
	GetHtmlVersion() string
	GetTitle() string
	CountHeadingLevels() map[string]int
	HasLoginForm() bool
	AnalyzeLinks() *LinkAnalysis
}
