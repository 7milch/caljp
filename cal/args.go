package cal

import "time"

// Options is the fully resolved command line.
type Options struct {
	First                          time.Time // first day of the first month shown
	Count                          int       // number of months
	Monday, Julian, Weeks, NoToday bool
	Color, Style                   string
	Refresh, Version, Help         bool
}

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
