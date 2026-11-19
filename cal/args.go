package cal

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const Usage = "usage: caljp [-13hjMSwy] [-A n] [-B n] [-m month] [--color when] [--style spec] [--refresh] [[month] year]"

// Options is the fully resolved command line.
type Options struct {
	First                          time.Time // first day of the first month shown
	Count                          int       // number of months
	Monday, Julian, Weeks, NoToday bool
	Color, Style                   string
	Refresh, Version, Help         bool
}

// unsupported lists cal options that caljp recognises but does not implement.
var unsupported = map[byte]string{
	'J': "Julian calendar", 'e': "date of Easter", 'o': "date of Orthodox Easter",
	'p': "country codes", 's': "country code", 'W': "first week of the year",
	'b': "oldstyle ncal format", 'C': "cal mode", 'N': "ncal mode",
	'd': "debug current date", 'H': "debug highlight date",
}

var errRange = fmt.Errorf("months outside the years 1..9999")

type unsupportedError string

func (e unsupportedError) Error() string { return string(e) }

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// Parse reads BSD cal style arguments; now supplies the default month and year.
func Parse(args []string, now time.Time) (Options, error) {
	o := Options{Color: "auto"}
	span, before, after, monthArg := byte(0), 0, 0, ""
	var pos []string
	var err error

	for i := 0; i < len(args) && err == nil; i++ {
		a := args[i]
		value := func(name, inline string, attached bool) string {
			if attached {
				return inline
			}
			if i++; i < len(args) {
				return args[i]
			}
			err = fmt.Errorf("option %s requires an argument", name)
			return ""
		}
		switch {
		case a == "--":
			pos, i = append(pos, args[i+1:]...), len(args)
		case strings.HasPrefix(a, "--"):
			name, inline, found := strings.Cut(a, "=")
			if found && name != "--color" && name != "--style" {
				err = fmt.Errorf("option %s takes no argument", name)
			}
			switch name {
			case "--refresh":
				o.Refresh = true
			case "--version":
				o.Version = true
			case "--help":
				o.Help = true
			case "--color":
				o.Color = value(name, inline, found)
			case "--style":
				o.Style = value(name, inline, found)
			default:
				err = fmt.Errorf("unknown option %s", name)
			}
		case len(a) > 1 && a[0] == '-':
			for j := 1; j < len(a) && err == nil; j++ {
				c, name := a[j], "-"+a[j:j+1]
				switch c {
				case '1', '3', 'y':
					span = c
				case 'h':
					o.NoToday = true
				case 'j':
					o.Julian = true
				case 'w':
					o.Weeks = true
				case 'M', 'S':
					o.Monday = c == 'M'
				case 'A', 'B', 'm':
					v := value(name, a[j+1:], j+1 < len(a))
					j = len(a)
					switch {
					case err != nil:
					case c == 'm':
						monthArg = v
					case c == 'A':
						after, err = strconv.Atoi(v)
					default:
						before, err = strconv.Atoi(v)
					}
					if _, bad := err.(*strconv.NumError); bad {
						err = fmt.Errorf("invalid number %q for %s", v, name)
					}
				default:
					if what, ok := unsupported[c]; ok {
						err = unsupportedError(fmt.Sprintf("%s (%s) is a cal option that caljp does not support", name, what))
					} else {
						err = fmt.Errorf("unknown option %s", name)
					}
				}
			}
		default:
			pos = append(pos, a)
		}
	}
	if err != nil {
		return o, err
	}
	if o.Color != "auto" && o.Color != "always" && o.Color != "never" {
		return o, fmt.Errorf("invalid --color value %q (auto, always or never)", o.Color)
	}

	first, count, err := months(pos, monthArg, span, now)
	if err != nil {
		return o, err
	}
	// Month numbers counted from January of year 0 keep the range check free of overflow.
	const limit = 10000 * 12
	if before < -limit || before > limit || after < -limit || after > limit {
		return o, errRange
	}
	start := first.Year()*12 + int(first.Month()) - 1 - before
	o.First, o.Count = first.AddDate(0, -before, 0), count+before+after
	switch {
	case o.Count < 1:
		return o, fmt.Errorf("no months to show")
	case start < 12 || start+o.Count > limit:
		return o, errRange
	}
	return o, nil
}
