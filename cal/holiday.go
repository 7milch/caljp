package cal

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HolidayURL is the official list published by the Cabinet Office of Japan.
const HolidayURL = "https://www8.cao.go.jp/chosei/shukujitsu/syukujitsu.csv"

// minHolidays rejects downloads that are not the real list.
const minHolidays = 100

//go:embed holidays.csv
var embedded []byte

// Holidays is the set of holiday dates, keyed by UTC midnight.
type Holidays map[time.Time]bool

// parseHolidays reads only the date column, so the encoding of the names does not matter.
func parseHolidays(r io.Reader) Holidays {
	h := Holidays{}
	for s := bufio.NewScanner(r); s.Scan(); {
		field, _, _ := strings.Cut(s.Text(), ",")
		if d, err := time.Parse("2006/1/2", field); err == nil {
			h[d] = true
		}
	}
	return h
}

// Range returns the earliest and latest holiday.
func (h Holidays) Range() (first, last time.Time) {
	for d := range h {
		if first.IsZero() || d.Before(first) {
			first = d
		}
		if d.After(last) {
			last = d
		}
	}
	return first, last
}

// CachePath is where --refresh stores the list; empty if the OS has no cache directory.
func CachePath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "caljp", "holidays.csv")
}

// LoadHolidays picks the embedded list or the cache, whichever reaches further into the future.
func LoadHolidays(cachePath string) Holidays {
	h := parseHolidays(bytes.NewReader(embedded))
	if data, err := os.ReadFile(cachePath); err == nil {
		cached := parseHolidays(bytes.NewReader(data))
		_, a := h.Range()
		if _, b := cached.Range(); len(cached) > 0 && !b.Before(a) {
			return cached
		}
	}
	return h
}

// Refresh downloads the list and replaces the cache only if it looks valid.
func Refresh(url, cachePath string) (Holidays, error) {
	if cachePath == "" {
		return nil, fmt.Errorf("no cache directory on this system")
	}
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	h := parseHolidays(bytes.NewReader(data))
	if resp.StatusCode != http.StatusOK || len(h) < minHolidays {
		return nil, fmt.Errorf("%s: %s, %d holidays found (want at least %d)", url, resp.Status, len(h), minHolidays)
	}

	// Write beside the target so the rename is atomic.
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return nil, err
	}
	tmp := cachePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return nil, err
	}
	return h, os.Rename(tmp, cachePath)
}
