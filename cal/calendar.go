package cal

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const gap = "  "

var dayNames = [...]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}

// Render draws the months selected by o.
func Render(o Options, p Palette) string {
	perRow := 3
	if o.Julian {
		perRow = 2
	}
	wholeYear := o.Count == 12 && o.First.Month() == time.January

	var out []string
	if wholeYear {
		total := perRow*o.width() + (perRow-1)*len(gap)
		out = append(out, center(strconv.Itoa(o.First.Year()), total), "")
	}
	for i := 0; i < o.Count; i += perRow {
		if i > 0 {
			out = append(out, "")
		}
		var blocks [][]string
		for j := i; j < min(i+perRow, o.Count); j++ {
			first := o.First.AddDate(0, j, 0)
			title := first.Month().String()
			if !wholeYear {
				title += " " + strconv.Itoa(first.Year())
			}
			blocks = append(blocks, block(o, p, first, title))
		}
		out = append(out, sideBySide(blocks, o.width())...)
	}

	var b strings.Builder
	for _, line := range out {
		b.WriteString(strings.TrimRight(line, " ") + "\n")
	}
	return b.String()
}

func (o Options) cell() int {
	if o.Julian {
		return 3
	}
	return 2
}

func (o Options) width() int {
	w := 7*o.cell() + 6
	if o.Weeks {
		w += 3
	}
	return w
}

// block returns one month as lines of equal visible width.
func block(o Options, p Palette, first time.Time, title string) []string {
	cell, start, margin := o.cell(), time.Sunday, ""
	if o.Monday {
		start = time.Monday
	}
	if o.Weeks {
		margin = "   "
	}
	head := make([]string, 7)
	for i := range head {
		wd := (start + time.Weekday(i)) % 7
		head[i] = strings.Repeat(" ", cell-2) + p.Head(dayNames[wd], wd)
	}
	lines := []string{
		margin + center(title, 7*cell+6),
		margin + strings.Join(head, " "),
	}

	d := first.AddDate(0, 0, -int((first.Weekday()-start+7)%7))
	for d.Before(first.AddDate(0, 1, 0)) {
		week, monday := make([]string, 7), d.AddDate(0, 0, int(time.Monday-start))
		for i := range week {
			week[i] = strings.Repeat(" ", cell)
			if d.Month() == first.Month() {
				n := d.Day()
				if o.Julian {
					n = d.YearDay()
				}
				s := strconv.Itoa(n)
				week[i] = strings.Repeat(" ", cell-len(s)) + p.Day(s, d)
			}
			d = d.AddDate(0, 0, 1)
		}
		line := strings.Join(week, " ")
		if o.Weeks {
			_, n := monday.ISOWeek()
			line = fmt.Sprintf("%2d ", n) + line
		}
		lines = append(lines, line)
	}
	return lines
}

func sideBySide(blocks [][]string, width int) []string {
	height := 0
	for _, b := range blocks {
		height = max(height, len(b))
	}
	lines := make([]string, height)
	for i := range lines {
		row := make([]string, len(blocks))
		for j, b := range blocks {
			row[j] = strings.Repeat(" ", width)
			if i < len(b) {
				row[j] = b[i]
			}
		}
		lines[i] = strings.Join(row, gap)
	}
	return lines
}

func center(s string, width int) string {
	left := (width - len(s)) / 2
	return fmt.Sprintf("%*s%-*s", left, "", width-left, s)
}
