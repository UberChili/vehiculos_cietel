package main

import "unicode"

// helper function to capitalize strings
func CapitalizeFirst(text string) string {
	if text == "" {
		return ""
	}
	runes := []rune(text)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
