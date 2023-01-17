package opencypher

import (
	"time"
)

// copied from Neo4j driver

type (
	Time          time.Time // Time since start of day with timezone information
	Date          time.Time // Date value, without a time zone and time related components.
	LocalTime     time.Time // Time since start of day in local timezone
	LocalDateTime time.Time // Date and time in local timezone
)

// Time casts LocalDateTime to time.Time
func (t LocalDateTime) Time() time.Time {
	return time.Time(t)
}

func (t LocalDateTime) String() string {
	return t.Time().Format("2006-01-02T15:04:05")
}

// Time casts LocalTime to time.Time
func (t LocalTime) Time() time.Time {
	return time.Time(t)
}

func (t LocalTime) String() string {
	return t.Time().Format("15:04:05")
}

// Time casts Date to time.Time
func (t Date) Time() time.Time {
	return time.Time(t)
}

func (t Date) String() string {
	return t.Time().Format("2006-01-02")
}

// Time casts Time to time.Time
func (t Time) Time() time.Time {
	return time.Time(t)
}

func (t Time) String() string {
	return t.Time().Format("15:04:05")
}

// NewDate creates a Date
// Hour, minute, second and nanoseconds are set to zero and location is set to UTC.
func NewDate(t time.Time) Date {
	y, m, d := t.Date()
	t = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return Date(t)
}

// NewLocalTime creates a LocalTime from time.Time.
// Year, month and day are set to zero and location is set to local.
func NewLocalTime(t time.Time) LocalTime {
	t = time.Date(0, 0, 0, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
	return LocalTime(t)
}

// NewLocalDateTime creates a LocalDateTime from time.Time.
func NewLocalDateTime(t time.Time) LocalDateTime {
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
	return LocalDateTime(t)
}
