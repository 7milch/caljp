package cal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// csvBody makes n consecutive "holidays" starting on January 1 of year.
func csvBody(year, n int) string {
	var b strings.Builder
	b.WriteString("date,name\r\n")
	for i := range n {
		b.WriteString(date(year, time.January, 1+i).Format("2006/1/2") + ",x\r\n")
	}
	return b.String()
}

func writeCache(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "caljp", "holidays.csv")
	os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseHolidays(t *testing.T) {
	h := parseHolidays(strings.NewReader("header,x\n2026/11/3,a\n2026/11/23,b\nbroken\n\n2026/1/1,\x8c\xb3\x93\xfa\r\n"))
	if len(h) != 3 || !h[date(2026, time.November, 3)] || !h[date(2026, time.January, 1)] {
		t.Errorf("parsed %v", h)
	}
	first, last := h.Range()
	if first != date(2026, time.January, 1) || last != date(2026, time.November, 23) {
		t.Errorf("range %v..%v", first, last)
	}
}

func TestEmbedded(t *testing.T) {
	h := LoadHolidays("")
	if len(h) < 900 || !h[date(2026, time.November, 3)] || !h[date(1955, time.January, 1)] {
		t.Errorf("embedded data looks wrong: %d days", len(h))
	}
	if h[date(2026, time.November, 4)] {
		t.Error("2026-11-04 is not a holiday")
	}
}

func TestLoadPrefersNewer(t *testing.T) {
	_, embeddedLast := LoadHolidays("").Range()
	future, past := date(embeddedLast.Year()+5, time.January, 1), date(1900, time.January, 1)

	if h := LoadHolidays(writeCache(t, csvBody(future.Year(), 3))); !h[future] || len(h) != 3 {
		t.Error("a cache with a later last date must win")
	}
	if h := LoadHolidays(writeCache(t, csvBody(1900, 3))); h[past] || len(h) < 900 {
		t.Error("a cache with an earlier last date must lose")
	}
	same := "x,y\n" + embeddedLast.Format("2006/1/2") + ",only\n"
	if h := LoadHolidays(writeCache(t, same)); len(h) != 1 {
		t.Error("on a tie the cache must win")
	}
	for name, path := range map[string]string{
		"empty":   writeCache(t, ""),
		"garbage": writeCache(t, "\x00\xff not a csv"),
		"missing": filepath.Join(t.TempDir(), "nope.csv"),
		"dir":     t.TempDir(),
	} {
		if h := LoadHolidays(path); len(h) < 900 {
			t.Errorf("%s cache: expected embedded data, got %d days", name, len(h))
		}
	}
}

func TestRefresh(t *testing.T) {
	body, status := csvBody(2030, 200), http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}))
	defer srv.Close()

	path := filepath.Join(t.TempDir(), "new", "dir", "holidays.csv")
	h, err := Refresh(srv.URL, path)
	if err != nil || len(h) != 200 || !h[date(2030, time.January, 1)] {
		t.Fatalf("refresh: %d days, err %v", len(h), err)
	}
	saved, _ := os.ReadFile(path)
	if string(saved) != body {
		t.Error("cache file does not hold the downloaded body")
	}
	if entries, _ := os.ReadDir(filepath.Dir(path)); len(entries) != 1 {
		t.Errorf("temporary files left behind: %v", entries)
	}

	failures := map[string]func(){
		"http 500":      func() { status = http.StatusInternalServerError },
		"too few dates": func() { status, body = http.StatusOK, csvBody(2031, 99) },
		"not a csv":     func() { status, body = http.StatusOK, "<html>moved</html>" },
	}
	for name, setup := range failures {
		setup()
		if _, err := Refresh(srv.URL, path); err == nil {
			t.Errorf("%s: expected an error", name)
		}
		if now, _ := os.ReadFile(path); string(now) != string(saved) {
			t.Errorf("%s: cache was modified", name)
		}
	}

	status, body = http.StatusOK, csvBody(2030, 200)+strings.Repeat("x", maxBody)
	if _, err := Refresh(srv.URL, path); err == nil {
		t.Error("oversized body: expected an error")
	}
	body = csvBody(2030, 200)
	blocked := filepath.Join(t.TempDir(), "holidays.csv")
	os.MkdirAll(filepath.Join(blocked, "occupied"), 0o755)
	if _, err := Refresh(srv.URL, blocked); err == nil {
		t.Error("cache path is a directory: expected an error")
	}
	if entries, _ := os.ReadDir(filepath.Dir(blocked)); len(entries) != 1 {
		t.Errorf("temporary files left behind after a failed rename: %v", entries)
	}

	srv.Close()
	if _, err := Refresh(srv.URL, path); err == nil {
		t.Error("unreachable server: expected an error")
	}
	if _, err := Refresh(srv.URL, ""); err == nil {
		t.Error("empty cache path: expected an error")
	}
}

func TestCachePath(t *testing.T) {
	if got, want := CachePath(), filepath.Join("caljp", "holidays.csv"); !strings.HasSuffix(got, want) {
		t.Errorf("CachePath() = %q, want suffix %q", got, want)
	}
}
