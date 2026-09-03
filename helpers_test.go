package main

import "testing"

func TestCapitalizeFirst(t *testing.T) {
	cases := map[string]string{
		"":                  "",
		"a":                 "A",
		"already":           "Already",
		"chevrolet spark":   "Chevrolet spark",
		"Already Uppercase": "Already Uppercase",
	}
	for in, want := range cases {
		if got := CapitalizeFirst(in); got != want {
			t.Errorf("CapitalizeFirst(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFormatDateDisplay(t *testing.T) {
	cases := []struct{ in, want string }{
		{"2026-09-02", "02-09-2026"},
		{"2026-01-15", "15-01-2026"},
		{"4023-05-04", "04-05-4023"}, // garbage-but-valid year still converts faithfully
		{"", ""},                     // no service yet: stays blank, doesn't render "01-01-0001" or similar
		{"not-a-date", "not-a-date"}, // unparseable input passes through rather than being hidden
	}
	for _, c := range cases {
		if got := FormatDateDisplay(c.in); got != c.want {
			t.Errorf("FormatDateDisplay(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
