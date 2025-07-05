package html

type LinkAnalysis struct {
	Internal     int `json:"internal"`
	External     int `json:"external"`
	Inaccessible int `json:"inaccessible"`
}
