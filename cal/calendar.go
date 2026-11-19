package cal

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var dayNames = [...]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}

// Render draws the months selected by o.
func Render(o Options, p Palette) string {
	title := fmt.Sprintf("%s %d", o.First.Month(), o.First.Year())
	var b strings.Builder
	for _, line := range block(o, p, o.First, title) {
		b.WriteString(strings.TrimRight(line, " ") + "\n")
	}
	return b.String()
}

// block returns one month as lines of equal visible width.
func block(o Options, p Palette, first time.Time, title string) []string {
	const cell = 2
	start := time.Sunday
	if o.Monday {
		start = time.Monday
	}
	head := make([]string, 7)
	for i := range head {
		wd := (start + time.Weekday(i)) % 7
		head[i] = p.Head(dayNames[wd], wd)
	}
	lines := []string{center(title, 7*cell+6), strings.Join(head, " ")}

	d := first.AddDate(0, 0, -int((first.Weekday()-start+7)%7))
	for d.Before(first.AddDate(0, 1, 0)) {
		week := make([]string, 7)
		for i := range week {
			week[i] = strings.Repeat(" ", cell)
			if d.Month() == first.Month() {
				s := strconv.Itoa(d.Day())
				week[i] = strings.Repeat(" ", cell-len(s)) + p.Day(s, d)
			}
			d = d.AddDate(0, 0, 1)
		}
		lines = append(lines, strings.Join(week, " "))
	}
	return lines
}

func center(s string, width int) string {
	left := (width - len(s)) / 2
	return fmt.Sprintf("%*s%-*s", left, "", width-left, s)
}
