package cal

import (
	"testing"
	"time"
)

var testHolidays = holidays{
	date(2026, time.November, 3):  true, // Tuesday
	date(2026, time.November, 22): true, // Sunday, made up for the test
	date(2026, time.November, 23): true, // Monday
}

func nov(d int) time.Time { return date(2026, time.November, d) }

func TestPaletteDefault(t *testing.T) {
	p, err := newPalette("", nov(19), testHolidays, true)
	if err != nil {
		t.Fatal(err)
	}
	days := map[int]string{
		1:  "\x1b[31m1\x1b[0m", // Sunday
		7:  "\x1b[34m7\x1b[0m", // Saturday
		3:  "\x1b[31m3\x1b[0m", // holiday
		4:  "4",                // plain weekday
		19: "\x1b[7m19\x1b[0m", // today
	}
	for d, want := range days {
		if got := p.day(nov(d).Format("2"), nov(d)); got != want {
			t.Errorf("day %d: %q, want %q", d, got, want)
		}
	}
	if got := p.head("Su", time.Sunday); got != "\x1b[31mSu\x1b[0m" {
		t.Errorf("sunday head %q", got)
	}
	if got := p.head("Mo", time.Monday); got != "Mo" {
		t.Errorf("monday head %q", got)
	}
}

func TestPaletteStyle(t *testing.T) {
	tests := []struct {
		style     string
		today     int
		highlight bool
		day       int
		want      string
	}{
		{"sat=cyan", 19, true, 7, "\x1b[36m7\x1b[0m"},
		{"sat=cyan", 19, true, 1, "\x1b[31m1\x1b[0m"},
		{"sun=none", 19, true, 1, "1"},
		{"holiday=none", 19, true, 3, "3"},
		{"holiday=none", 19, true, 22, "\x1b[31m22\x1b[0m"},
		{"holiday=none,sun=green", 19, true, 22, "\x1b[32m22\x1b[0m"},
		{"mon=green,holiday=yellow", 19, true, 23, "\x1b[33m23\x1b[0m"},
		{"mon=green,holiday=yellow", 19, true, 16, "\x1b[32m16\x1b[0m"},
		{"thu=magenta,fri=white,wed=black,tue=blue", 1, true, 5, "\x1b[35m5\x1b[0m"},
		{"", 19, false, 19, "19"},
		{"", 22, true, 22, "\x1b[7;31m22\x1b[0m"},
		{"", 22, false, 22, "\x1b[31m22\x1b[0m"},
	}
	for _, tt := range tests {
		p, err := newPalette(tt.style, nov(tt.today), testHolidays, tt.highlight)
		if err != nil {
			t.Errorf("%q: %v", tt.style, err)
			continue
		}
		if got := p.day(nov(tt.day).Format("2"), nov(tt.day)); got != tt.want {
			t.Errorf("%q day %d: %q, want %q", tt.style, tt.day, got, tt.want)
		}
	}
}

func TestPaletteErrors(t *testing.T) {
	for _, style := range []string{"foo=red", "sat=pink", "sat", "sat=", "=red", "sat=red,", "sat=red;sun=blue"} {
		if _, err := newPalette(style, nov(19), nil, true); err == nil {
			t.Errorf("%q: expected an error", style)
		}
	}
}

func TestZeroPalette(t *testing.T) {
	var p palette
	if p.day("1", nov(1)) != "1" || p.head("Su", time.Sunday) != "Su" {
		t.Error("zero palette must not decorate")
	}
}

func TestUseColor(t *testing.T) {
	tests := []struct {
		mode         string
		tty, noColor bool
		want         bool
	}{
		{"always", false, true, true},
		{"never", true, false, false},
		{"auto", true, false, true},
		{"auto", false, false, false},
		{"auto", true, true, false},
	}
	for _, tt := range tests {
		if got := useColor(tt.mode, tt.tty, tt.noColor); got != tt.want {
			t.Errorf("useColor(%q, %v, %v) = %v", tt.mode, tt.tty, tt.noColor, got)
		}
	}
}
