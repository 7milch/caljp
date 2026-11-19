package cal

import (
	"strings"
	"testing"
	"time"
)

func month(y int, m time.Month) Options {
	return Options{First: date(y, m, 1), Count: 1}
}

func checkShape(t *testing.T, out string) {
	t.Helper()
	if !strings.HasSuffix(out, "\n") || strings.HasSuffix(out, "\n\n") {
		t.Errorf("output must end with exactly one newline: %q", out)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasSuffix(line, " ") {
			t.Errorf("trailing space in %q", line)
		}
	}
}

func TestRenderMonth(t *testing.T) {
	nov := `   November 2026
Su Mo Tu We Th Fr Sa
 1  2  3  4  5  6  7
 8  9 10 11 12 13 14
15 16 17 18 19 20 21
22 23 24 25 26 27 28
29 30
`
	novMonday := `   November 2026
Mo Tu We Th Fr Sa Su
                   1
 2  3  4  5  6  7  8
 9 10 11 12 13 14 15
16 17 18 19 20 21 22
23 24 25 26 27 28 29
30
`
	monday := month(2026, time.November)
	monday.Monday = true

	if got := Render(month(2026, time.November), Palette{}); got != nov {
		t.Errorf("sunday start:\n%s\nwant:\n%s", got, nov)
	}
	if got := Render(monday, Palette{}); got != novMonday {
		t.Errorf("monday start:\n%s\nwant:\n%s", got, novMonday)
	}

	leap := Render(month(2024, time.February), Palette{})
	if !strings.HasSuffix(leap, "\n25 26 27 28 29\n") {
		t.Errorf("leap february:\n%s", leap)
	}

	four := Render(month(2026, time.February), Palette{})
	if n := strings.Count(four, "\n"); n != 6 {
		t.Errorf("february 2026 has %d lines, want 6:\n%s", n, four)
	}

	for _, out := range []string{nov, leap, four, Render(monday, Palette{})} {
		checkShape(t, out)
	}
}
