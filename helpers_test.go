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

func TestParseDisplayDate(t *testing.T) {
	t.Run("valid dd-mm-yyyy converts to ISO", func(t *testing.T) {
		got, err := ParseDisplayDate("02-09-2026")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "2026-09-02" {
			t.Errorf("got %q, want %q", got, "2026-09-02")
		}
	})

	t.Run("round-trips with FormatDateDisplay", func(t *testing.T) {
		iso := "2026-12-31"
		display := FormatDateDisplay(iso)
		back, err := ParseDisplayDate(display)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if back != iso {
			t.Errorf("round trip: got %q, want %q", back, iso)
		}
	})

	for _, bad := range []string{"2026-09-02", "31/02/2026", "not-a-date", ""} {
		t.Run("rejects "+bad, func(t *testing.T) {
			if _, err := ParseDisplayDate(bad); err == nil {
				t.Errorf("expected an error for %q", bad)
			}
		})
	}
}
