package cal

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// months resolves positional arguments, -m and -1/-3/-y into a first month and a count.
func months(pos []string, monthArg string, span byte, now time.Time) (time.Time, int, error) {
	year, month, err := now.Year(), now.Month(), error(nil)
	switch len(pos) {
	case 0:
	case 1:
		if m, e := parseMonth(pos[0]); e == nil && !isNumber(pos[0]) {
			month = m
		} else if year, err = parseYear(pos[0]); err == nil && monthArg == "" {
			if span == '1' || span == '3' {
				err = fmt.Errorf("-%c cannot be combined with a year alone; add a month", span)
			}
			span = 'y'
		}
	case 2:
		if month, err = parseMonth(pos[0]); err == nil {
			year, err = parseYear(pos[1])
		}
	default:
		err = fmt.Errorf("too many arguments")
	}
	if err == nil && monthArg != "" {
		// A trailing f or p on a numeric month means the following or previous year.
		if n := len(monthArg) - 1; n > 0 && isNumber(monthArg[:n]) {
			year += map[byte]int{'f': 1, 'p': -1}[monthArg[n]]
			monthArg = strings.TrimRight(monthArg, "fp")
		}
		month, err = parseMonth(monthArg)
	}
	if err != nil {
		return time.Time{}, 0, err
	}
	switch span {
	case '3':
		return date(year, month-1, 1), 3, nil
	case 'y':
		return date(year, time.January, 1), 12, nil
	}
	return date(year, month, 1), 1, nil
}

func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func parseYear(s string) (int, error) {
	y, err := strconv.Atoi(s)
	if err != nil || y < 1 || y > 9999 {
		return 0, fmt.Errorf("invalid year %q (1..9999)", s)
	}
	return y, nil
}

// parseMonth accepts 1..12, an English month name or its three-letter form.
func parseMonth(s string) (time.Month, error) {
	if n, err := strconv.Atoi(s); err == nil && n >= 1 && n <= 12 {
		return time.Month(n), nil
	}
	for m := time.January; m <= time.December; m++ {
		if name := m.String(); strings.EqualFold(s, name) || strings.EqualFold(s, name[:3]) {
			return m, nil
		}
	}
	return 0, fmt.Errorf("invalid month %q", s)
}
