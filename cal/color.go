package cal

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

const defaultStyle = "sun=red,sat=blue,holiday=red"

var (
	// styleKeys is indexed by time.Weekday; "holiday" comes last.
	styleKeys = []string{"sun", "mon", "tue", "wed", "thu", "fri", "sat", "holiday"}
	colors    = map[string]int{
		"none": 0, "black": 30, "red": 31, "green": 32, "yellow": 33,
		"blue": 34, "magenta": 35, "cyan": 36, "white": 37,
	}
)

// Palette decorates cells; the zero value leaves text unchanged.
type Palette struct {
	codes     map[string]int
	today     time.Time
	holidays  Holidays
	highlight bool
}

// NewPalette applies style on top of the defaults; highlight marks today in reverse video.
func NewPalette(style string, today time.Time, hol Holidays, highlight bool) (Palette, error) {
	p := Palette{map[string]int{}, today, hol, highlight}
	if style != "" {
		style = "," + style
	}
	for _, pair := range strings.Split(defaultStyle+style, ",") {
		key, name, _ := strings.Cut(pair, "=")
		code, ok := colors[name]
		if !ok || !slices.Contains(styleKeys, key) {
			return p, fmt.Errorf("invalid style %q (want e.g. sat=blue; colors: none black red green yellow blue magenta cyan white)", pair)
		}
		p.codes[key] = code
	}
	return p, nil
}

// UseColor decides whether to emit escape codes; tty and noColor only matter for "auto".
func UseColor(mode string, tty, noColor bool) bool {
	return mode == "always" || mode == "auto" && tty && !noColor
}

func (p Palette) Day(text string, d time.Time) string {
	code := p.codes[styleKeys[d.Weekday()]]
	if h := p.codes["holiday"]; h != 0 && p.holidays[d] {
		code = h
	}
	return paint(text, code, p.highlight && d.Equal(p.today))
}

func (p Palette) Head(text string, wd time.Weekday) string {
	return paint(text, p.codes[styleKeys[wd]], false)
}

func paint(text string, code int, reverse bool) string {
	var sgr []string
	if reverse {
		sgr = append(sgr, "7")
	}
	if code != 0 {
		sgr = append(sgr, strconv.Itoa(code))
	}
	if sgr == nil {
		return text
	}
	return "\x1b[" + strings.Join(sgr, ";") + "m" + text + "\x1b[0m"
}
