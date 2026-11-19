package cal

import (
	"errors"
	"fmt"
	"io"
	"time"
)

const help = usage + `

  -1, -3, -y      one month (default), three months, the whole year
  -A n, -B n      add n months after / before
  -m month        show this month (8f / 8p: next / previous year)
  -M, -S          weeks start on Monday / Sunday (default)
  -h              do not highlight today
  -j              show days of the year
  -w              show ISO week numbers
  --color when    auto (default), always or never
  --style spec    colors, e.g. sat=blue,sun=red,holiday=red
                  keys: sun mon tue wed thu fri sat holiday
                  colors: none black red green yellow blue magenta cyan white
  --refresh       download the latest Japanese holiday list and exit
  --version       print the version and exit
`

// Env is everything Run needs from the outside world, so tests can pin it.
type Env struct {
	Now     time.Time // today at UTC midnight
	TTY     bool      // stdout is a terminal
	NoColor bool      // NO_COLOR is set
	Cache   string    // holiday cache file
	URL     string    // holiday CSV source
	Version string
}

// Run executes caljp and returns the exit code.
func Run(args []string, stdout, stderr io.Writer, env Env) int {
	o, err := parse(args, env.Now)
	var p palette
	if err == nil {
		p, err = newPalette(o.Style, env.Now, loadHolidays(env.Cache), !o.NoToday)
	}
	if err != nil {
		fmt.Fprintf(stderr, "caljp: %v\n", err)
		if !errors.As(err, new(unsupportedError)) {
			fmt.Fprintln(stderr, usage)
		}
		return 2
	}

	switch {
	case o.Help:
		fmt.Fprint(stdout, help)
	case o.Version:
		fmt.Fprintf(stdout, "caljp %s\n", env.Version)
	case o.Refresh:
		h, err := refresh(env.URL, env.Cache)
		if err != nil {
			fmt.Fprintf(stderr, "caljp: %v\n", err)
			return 1
		}
		first, last := h.bounds()
		fmt.Fprintf(stdout, "holidays updated: %s..%s (%d days)\n",
			first.Format(time.DateOnly), last.Format(time.DateOnly), len(h))
	default:
		if !useColor(o.Color, env.TTY, env.NoColor) {
			p = palette{}
		}
		fmt.Fprint(stdout, render(o, p))
	}
	return 0
}
