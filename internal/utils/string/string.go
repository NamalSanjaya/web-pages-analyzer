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

func ContainsAnyPrefix(s string, prefixArr ...string) bool {
	for _, prefix := range prefixArr {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}
