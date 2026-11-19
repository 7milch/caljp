package cal

import "time"

// Holidays is the set of holiday dates, keyed by UTC midnight.
type Holidays map[time.Time]bool
