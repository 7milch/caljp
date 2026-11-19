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

type unsupportedError string

func (e unsupportedError) Error() string { return string(e) }

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// Parse reads BSD cal style arguments; now supplies the default month and year.
func Parse(args []string, now time.Time) (Options, error) {
	o := Options{Color: "auto"}
	span, before, after, monthArg := byte('1'), 0, 0, ""
	var pos []string
	var err error

	for i := 0; i < len(args) && err == nil; i++ {
		a := args[i]
		value := func(name, inline string) string {
			if inline != "" {
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
			name, inline, _ := strings.Cut(a, "=")
			switch name {
			case "--refresh":
				o.Refresh = true
			case "--version":
				o.Version = true
			case "--help":
				o.Help = true
			case "--color":
				o.Color = value(name, inline)
			case "--style":
				o.Style = value(name, inline)
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
					v := value(name, a[j+1:])
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

	o.First, o.Count, err = months(pos, monthArg, span, now)
	o.First = o.First.AddDate(0, -before, 0)
	if o.Count += before + after; err == nil && o.Count < 1 {
		err = fmt.Errorf("no months to show")
	}
	return o, err
}
