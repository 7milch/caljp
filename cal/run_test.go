package cal

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const november = `   November 2026
Su Mo Tu We Th Fr Sa
 1  2  3  4  5  6  7
 8  9 10 11 12 13 14
15 16 17 18 19 20 21
22 23 24 25 26 27 28
29 30
`

func run(t *testing.T, env Env, args string) (code int, stdout, stderr string) {
	t.Helper()
	if env.Cache == "" {
		env.Cache = filepath.Join(t.TempDir(), "holidays.csv")
	}
	env.Now, env.Version = now, "2026.11.0"
	var out, errOut bytes.Buffer
	code = Run(strings.Fields(args), &out, &errOut, env)
	return code, out.String(), errOut.String()
}

func TestRunColor(t *testing.T) {
	const today, holiday, sunday = "\x1b[7m19\x1b[0m", "\x1b[31m3\x1b[0m", "\x1b[31mSu\x1b[0m"
	tests := []struct {
		env     Env
		args    string
		with    []string
		without []string
	}{
		{Env{}, "", nil, []string{"\x1b"}},
		{Env{TTY: true}, "", []string{today, holiday, sunday}, nil},
		{Env{TTY: true, NoColor: true}, "", nil, []string{"\x1b"}},
		{Env{}, "--color=always", []string{today, holiday}, nil},
		{Env{NoColor: true}, "--color=always", []string{today}, nil},
		{Env{TTY: true}, "--color=never", nil, []string{"\x1b"}},
		{Env{TTY: true}, "-h", []string{holiday}, []string{"\x1b[7m"}},
		{Env{TTY: true}, "--style sun=none,holiday=green", []string{"\x1b[32m3\x1b[0m", "\x1b[32m23\x1b[0m"}, []string{"\x1b[31m"}},
		{Env{TTY: true}, "-3", []string{"\x1b[31m23\x1b[0m"}, nil},
	}
	for _, tt := range tests {
		code, out, errOut := run(t, tt.env, tt.args)
		if code != 0 || errOut != "" {
			t.Errorf("%q %+v: exit %d, stderr %q", tt.args, tt.env, code, errOut)
		}
		for _, s := range tt.with {
			if !strings.Contains(out, s) {
				t.Errorf("%q %+v: missing %q in %q", tt.args, tt.env, s, out)
			}
		}
		for _, s := range tt.without {
			if strings.Contains(out, s) {
				t.Errorf("%q %+v: unexpected %q in %q", tt.args, tt.env, s, out)
			}
		}
	}
	if _, out, _ := run(t, Env{}, ""); out != november {
		t.Errorf("plain output:\n%s", out)
	}
	// Stripping the escape codes must give back the plain calendar.
	_, colored, _ := run(t, Env{TTY: true}, "")
	if plain := regexp.MustCompile("\x1b\\[[0-9;]*m").ReplaceAllString(colored, ""); plain != november {
		t.Errorf("colored output does not align:\n%s", plain)
	}
}

func TestRunErrors(t *testing.T) {
	code, out, errOut := run(t, Env{}, "-J")
	if code != 2 || out != "" || errOut != "caljp: -J (Julian calendar) is a cal option that caljp does not support\n" {
		t.Errorf("-J: exit %d, stdout %q, stderr %q", code, out, errOut)
	}
	code, out, errOut = run(t, Env{}, "-x")
	if code != 2 || out != "" || errOut != "caljp: unknown option -x\n"+usage+"\n" {
		t.Errorf("-x: exit %d, stdout %q, stderr %q", code, out, errOut)
	}
	for _, args := range []string{"--style sat=pink", "13 2027", "-A", "--color=sometimes"} {
		if code, out, errOut := run(t, Env{}, args); code != 2 || out != "" || !strings.HasPrefix(errOut, "caljp: ") {
			t.Errorf("%q: exit %d, stdout %q, stderr %q", args, code, out, errOut)
		}
	}
}

func TestRunInfo(t *testing.T) {
	if code, out, _ := run(t, Env{}, "--version"); code != 0 || out != "caljp 2026.11.0\n" {
		t.Errorf("--version: exit %d, %q", code, out)
	}
	code, out, errOut := run(t, Env{}, "--help")
	if code != 0 || errOut != "" || !strings.HasPrefix(out, usage+"\n") {
		t.Errorf("--help: exit %d, %q", code, out)
	}
	for _, s := range []string{"-M", "--style", "--refresh", "--color"} {
		if !strings.Contains(out, s) {
			t.Errorf("--help does not mention %s", s)
		}
	}
}

func TestRunRefresh(t *testing.T) {
	status := http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		fmt.Fprint(w, csvBody(2030, 200))
	}))
	defer srv.Close()
	cache := filepath.Join(t.TempDir(), "holidays.csv")

	code, out, errOut := run(t, Env{URL: srv.URL, Cache: cache}, "--refresh")
	if code != 0 || errOut != "" || out != "holidays updated: 2030-01-01..2030-07-19 (200 days)\n" {
		t.Errorf("--refresh: exit %d, stdout %q, stderr %q", code, out, errOut)
	}
	// The refreshed cache reaches further than the embedded list, so it is used.
	if _, out, _ := run(t, Env{Cache: cache}, "--color=always 1 2030"); !strings.Contains(out, "\x1b[31m2\x1b[0m") {
		t.Errorf("refreshed holidays are not used:\n%q", out)
	}

	status = http.StatusInternalServerError
	code, out, errOut = run(t, Env{URL: srv.URL, Cache: cache}, "--refresh")
	if code != 1 || out != "" || !strings.HasPrefix(errOut, "caljp: ") || !strings.Contains(errOut, "500") {
		t.Errorf("failed --refresh: exit %d, stdout %q, stderr %q", code, out, errOut)
	}
}
