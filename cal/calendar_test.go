package cal

import (
	"strings"
	"testing"
	"time"
)

func month(y int, m time.Month) options {
	return options{First: date(y, m, 1), Count: 1}
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

	if got := render(month(2026, time.November), palette{}); got != nov {
		t.Errorf("sunday start:\n%s\nwant:\n%s", got, nov)
	}
	if got := render(monday, palette{}); got != novMonday {
		t.Errorf("monday start:\n%s\nwant:\n%s", got, novMonday)
	}

	leap := render(month(2024, time.February), palette{})
	if !strings.HasSuffix(leap, "\n25 26 27 28 29\n") {
		t.Errorf("leap february:\n%s", leap)
	}

	four := render(month(2026, time.February), palette{})
	if n := strings.Count(four, "\n"); n != 6 {
		t.Errorf("february 2026 has %d lines, want 6:\n%s", n, four)
	}

	for _, out := range []string{nov, leap, four, render(monday, palette{})} {
		checkShape(t, out)
	}
}

func lines(o options) []string {
	return strings.Split(strings.TrimSuffix(render(o, palette{}), "\n"), "\n")
}

func TestRenderThree(t *testing.T) {
	got := lines(options{First: date(2026, time.October, 1), Count: 3})
	for title, col := range map[string]int{"October 2026": 4, "November 2026": 25, "December 2026": 47} {
		if i := strings.Index(got[0], title); i != col {
			t.Errorf("%s at column %d, want %d: %q", title, i, col, got[0])
		}
	}
	if want := "Su Mo Tu We Th Fr Sa  Su Mo Tu We Th Fr Sa  Su Mo Tu We Th Fr Sa"; got[1] != want {
		t.Errorf("header %q", got[1])
	}
	want := strings.Repeat(" ", 12) + " 1  2  3" + "  " + " 1  2  3  4  5  6  7" + "  " + strings.Repeat(" ", 6) + " 1  2  3  4  5"
	if got[2] != want {
		t.Errorf("first week\n%q\n%q", got[2], want)
	}
	if len(got) != 7 {
		t.Errorf("%d lines, want 7", len(got))
	}

	// August 2026 needs six week rows, its neighbours five.
	got = lines(options{First: date(2026, time.July, 1), Count: 3})
	if len(got) != 8 || got[7] != strings.Repeat(" ", 22)+"30 31" {
		t.Errorf("uneven months:\n%s", strings.Join(got, "\n"))
	}
	checkShape(t, render(options{First: date(2026, time.July, 1), Count: 3}, palette{}))
}

func TestRenderYear(t *testing.T) {
	out := render(options{First: date(2027, time.January, 1), Count: 12}, palette{})
	got := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if got[0] != strings.Repeat(" ", 30)+"2027" || got[1] != "" {
		t.Errorf("year title %q %q", got[0], got[1])
	}
	for title, col := range map[string]int{"January": 6, "February": 28, "March": 51} {
		if i := strings.Index(got[2], title); i != col {
			t.Errorf("%s at column %d, want %d: %q", title, i, col, got[2])
		}
	}
	if n := strings.Count(out, "2027"); n != 1 {
		t.Errorf("year appears %d times, want 1", n)
	}
	if n := strings.Count(out, "\n\n"); n != 4 {
		t.Errorf("%d blank lines, want 4", n)
	}
	if !strings.Contains(out, "October") || !strings.Contains(out, "December") {
		t.Errorf("missing months:\n%s", out)
	}
	checkShape(t, out)
}

func TestRenderSpanYears(t *testing.T) {
	got := lines(options{First: date(2026, time.December, 1), Count: 3})
	for _, title := range []string{"December 2026", "January 2027", "February 2027"} {
		if !strings.Contains(got[0], title) {
			t.Errorf("missing %q in %q", title, got[0])
		}
	}
	// Twelve months that are not one calendar year get no year title.
	got = lines(options{First: date(2026, time.November, 1), Count: 12})
	if !strings.Contains(got[0], "November 2026") {
		t.Errorf("first line %q", got[0])
	}
}

func TestRenderJulian(t *testing.T) {
	got := lines(options{First: date(2026, time.February, 1), Count: 1, Julian: true})
	if got[0] != strings.Repeat(" ", 7)+"February 2026" {
		t.Errorf("title %q", got[0])
	}
	if got[1] != " Su  Mo  Tu  We  Th  Fr  Sa" {
		t.Errorf("header %q", got[1])
	}
	if got[2] != " 32  33  34  35  36  37  38" {
		t.Errorf("first week %q", got[2])
	}
	got = lines(options{First: date(2026, time.February, 1), Count: 3, Julian: true})
	if !strings.Contains(got[0], "March 2026") || strings.Contains(got[0], "April") {
		t.Errorf("two months per row expected: %q", got[0])
	}
}

func TestRenderWeeks(t *testing.T) {
	got := lines(options{First: date(2026, time.November, 1), Count: 1, Monday: true, Weeks: true})
	want := []string{
		"      November 2026",
		"   Mo Tu We Th Fr Sa Su",
		"44                    1",
		"45  2  3  4  5  6  7  8",
		"46  9 10 11 12 13 14 15",
		"47 16 17 18 19 20 21 22",
		"48 23 24 25 26 27 28 29",
		"49 30",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

// A lone Sunday at the end of a Sunday-first month belongs to the next ISO week.
func TestRenderWeeksSundayStart(t *testing.T) {
	got := lines(options{First: date(2026, time.May, 1), Count: 1, Weeks: true})
	want := []string{
		"         May 2026",
		"   Su Mo Tu We Th Fr Sa",
		"18                 1  2",
		"19  3  4  5  6  7  8  9",
		"20 10 11 12 13 14 15 16",
		"21 17 18 19 20 21 22 23",
		"22 24 25 26 27 28 29 30",
		"23 31",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}
