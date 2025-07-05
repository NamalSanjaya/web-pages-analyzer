package string

import "strings"

func ContainsAnySubstring(s string, substrArr ...string) bool {
	for _, substr := range substrArr {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}
