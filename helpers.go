package main

import (
	"time"
	"unicode"
)

// helper function to capitalize strings
func CapitalizeFirst(text string) string {
	if text == "" {
		return ""
	}
	runes := []rune(text)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// FormatDateDisplay converts a date stored as ISO 8601 (YYYY-MM-DD) into
// dd-mm-yyyy, the format shown to users everywhere in the UI. Dates stay
// ISO in the database (so MAX()/ORDER BY sort correctly); this only
// affects display. Falls back to the original string unchanged if it
// isn't a valid ISO date — covers the empty "no service yet" case, and
// avoids ever rendering a blank date as something misleading.
func FormatDateDisplay(iso string) string {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return iso
	}
	return t.Format("02-01-2006")
}

// ParseDisplayDate is FormatDateDisplay's inverse: it converts a date
// entered by the user as dd-mm-yyyy (the format the record form uses,
// matching what's shown everywhere else) into ISO 8601 for storage.
func ParseDisplayDate(display string) (string, error) {
	t, err := time.Parse("02-01-2006", display)
	if err != nil {
		return "", err
	}
	return t.Format("2006-01-02"), nil
}
