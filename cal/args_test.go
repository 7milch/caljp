package cal

import (
	"strings"
	"testing"
	"time"
)

var now = date(2026, time.November, 19)

func TestParse(t *testing.T) {
	base := Options{Color: "auto"}
	with := func(edit func(*Options)) Options {
		o := base
		edit(&o)
		return o
	}
	tests := []struct {
		args  string
		first string
		count int
		flags Options // First and Count are ignored
	}{
		{"", "2026-11", 1, base},
		{"2027", "2027-01", 12, base},
		{"3 2027", "2027-03", 1, base},
		{"march", "2026-03", 1, base},
		{"Mar", "2026-03", 1, base},
		{"sep 2030", "2030-09", 1, base},
		{"-1", "2026-11", 1, base},
		{"-3", "2026-10", 3, base},
		{"-y", "2026-01", 12, base},
		{"-3 -y", "2026-01", 12, base},
		{"-y -3", "2026-10", 3, base},
		{"-3M", "2026-10", 3, with(func(o *Options) { o.Monday = true })},
		{"-M -S", "2026-11", 1, base},
		{"-S -M", "2026-11", 1, with(func(o *Options) { o.Monday = true })},
		{"-A2", "2026-11", 3, base},
		{"-A 2", "2026-11", 3, base},
		{"-B 1 -A 1", "2026-10", 3, base},
		{"-y -B-6", "2026-07", 6, base},
		{"-A -1 -3", "2026-10", 2, base},
		{"-A11", "2026-11", 12, base},
		{"-m 8", "2026-08", 1, base},
		{"-m8", "2026-08", 1, base},
		{"-m 8f", "2027-08", 1, base},
		{"-m 8p", "2025-08", 1, base},
		{"-m aug 2030", "2030-08", 1, base},
		{"-m sep", "2026-09", 1, base},
		{"-hjw", "2026-11", 1, with(func(o *Options) { o.NoToday, o.Julian, o.Weeks = true, true, true })},
		{"--color=never", "2026-11", 1, with(func(o *Options) { o.Color = "never" })},
		{"--color always", "2026-11", 1, with(func(o *Options) { o.Color = "always" })},
		{"--style sat=cyan", "2026-11", 1, with(func(o *Options) { o.Style = "sat=cyan" })},
		{"--style=sat=cyan,sun=none", "2026-11", 1, with(func(o *Options) { o.Style = "sat=cyan,sun=none" })},
		{"--refresh", "2026-11", 1, with(func(o *Options) { o.Refresh = true })},
		{"--version", "2026-11", 1, with(func(o *Options) { o.Version = true })},
		{"--help", "2026-11", 1, with(func(o *Options) { o.Help = true })},
		{"-- 2027", "2027-01", 12, base},
		{"-3 -m 8 2027", "2027-07", 3, base},
		{"-3 -y 2027", "2027-01", 12, base},
		{"--style= 2027", "2027-01", 12, base},
	}
	for _, tt := range tests {
		got, err := Parse(strings.Fields(tt.args), now)
		if err != nil {
			t.Errorf("%q: unexpected error: %v", tt.args, err)
			continue
		}
		if f := got.First.Format("2006-01"); f != tt.first || got.Count != tt.count {
			t.Errorf("%q: range %s x%d, want %s x%d", tt.args, f, got.Count, tt.first, tt.count)
		}
		got.First, got.Count = time.Time{}, 0
		if got != tt.flags {
			t.Errorf("%q: flags %+v, want %+v", tt.args, got, tt.flags)
		}
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct{ args, want string }{
		{"-J", "-J (Julian calendar) is a cal option that caljp does not support"},
		{"-e", "does not support"},
		{"-o", "does not support"},
		{"-p", "does not support"},
		{"-s jp", "does not support"},
		{"-W 4", "does not support"},
		{"-b", "does not support"},
		{"-C", "does not support"},
		{"-N", "does not support"},
		{"-d 2026-01", "does not support"},
		{"-H 2026-01-01", "does not support"},
		{"-3J", "does not support"},
		{"-x", "unknown option -x"},
		{"--bogus", "unknown option --bogus"},
		{"13 2027", "month"},
		{"0 2027", "month"},
		{"0", "year"},
		{"10000", "year"},
		{"3 0", "year"},
		{"foo", "foo"},
		{"3 foo", "foo"},
		{"1 2 3", "too many"},
		{"-A x", "x"},
		{"-A", "-A"},
		{"-m", "-m"},
		{"-m 13", "month"},
		{"-m 0f", "month"},
		{"--style", "--style"},
		{"--color", "--color"},
		{"--color=sometimes", "sometimes"},
		{"-B -1", "month"},
		{"-", "-"},
		{"-3 2027", "-3 cannot be combined with a year alone"},
		{"2027 -3", "-3 cannot be combined with a year alone"},
		{"-1 2027", "-1 cannot be combined with a year alone"},
		{"--color= never", `invalid --color value ""`},
		{"--refresh=x", "--refresh takes no argument"},
		{"--help=", "--help takes no argument"},
		{"--version=1", "--version takes no argument"},
	}
	for _, tt := range tests {
		_, err := Parse(strings.Fields(tt.args), now)
		if err == nil {
			t.Errorf("%q: expected an error", tt.args)
		} else if !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%q: error %q does not mention %q", tt.args, err, tt.want)
		}
	}
	if _, err := Parse([]string{"-J"}, now); err.Error() != "-J (Julian calendar) is a cal option that caljp does not support" {
		t.Errorf("-J message: %q", err)
	}
}

func TestParseRange(t *testing.T) {
	for _, args := range []string{
		"-A 30000000", "-B 99999", "-3 1 1", "-m 1p 1", "-m 8f 9999", "-3 12 9999",
		"-B 4000000000000 -A -3999999999999", "-A 9223372036854775807",
	} {
		if _, err := Parse(strings.Fields(args), now); err == nil || !strings.Contains(err.Error(), "1..9999") {
			t.Errorf("%q: want a range error, got %v", args, err)
		}
	}
	for _, args := range []string{"1 1", "12 9999", "-3 2 1", "-y 9999", "-B 24310"} {
		if _, err := Parse(strings.Fields(args), now); err != nil {
			t.Errorf("%q: unexpected error %v", args, err)
		}
	}
}
