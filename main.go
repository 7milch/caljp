package main

import (
	"os"
	"time"

	"github.com/7milch/caljp/cal"
)

// version is set by the release build.
var version = "dev"

func main() {
	now := time.Now()
	stat, err := os.Stdout.Stat()
	os.Exit(cal.Run(os.Args[1:], os.Stdout, os.Stderr, cal.Env{
		Now:     time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
		TTY:     err == nil && stat.Mode()&os.ModeCharDevice != 0,
		NoColor: os.Getenv("NO_COLOR") != "",
		Cache:   cal.CachePath(),
		URL:     cal.HolidayURL,
		Version: version,
	}))
}
