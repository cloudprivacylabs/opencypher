package opencypher

import (
	"fmt"
	"time"
)

// Duration represents a duration in opencypher. Copied from neo4j driver.
type Duration struct {
	Months  int64
	Days    int64
	Seconds int64
	Nanos   int
}

func (d Duration) String() string {
	sign := ""
	if d.Seconds < 0 && d.Nanos > 0 {
		d.Seconds++
		d.Nanos = int(time.Second) - d.Nanos

		if d.Seconds == 0 {
			sign = "-"
		}
	}

	timePart := ""
	if d.Nanos == 0 {
		timePart = fmt.Sprintf("%s%d", sign, d.Seconds)
	} else {
		timePart = fmt.Sprintf("%s%d.%09d", sign, d.Seconds, d.Nanos)
	}

	return fmt.Sprintf("P%dM%dDT%sS", d.Months, d.Days, timePart)
}

func (d1 Duration) Equal(d2 Duration) bool {
	return d1.Months == d2.Months && d1.Days == d2.Days && d1.Seconds == d2.Seconds && d1.Nanos == d2.Nanos
}
