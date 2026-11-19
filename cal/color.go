package cal

import "time"

// Palette decorates cells; the zero value leaves text unchanged.
type Palette struct{}

func (p Palette) Day(text string, d time.Time) string { return text }

func (p Palette) Head(text string, wd time.Weekday) string { return text }
