package xtime

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Common durations
const (
	Nanosecond  = time.Nanosecond
	Microsecond = time.Microsecond
	Millisecond = time.Millisecond
	Second      = time.Second
	Minute      = time.Minute
	Hour        = time.Hour
	Day         = 24 * Hour
	Week        = 7 * Day
	Month       = 30 * Day  // Approximate
	Year        = 365 * Day // Approximate
)

// IsLeapYear checks if the given year is a leap year
func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// DaysInMonth returns the number of days in the given month and year
func DaysInMonth(year, month int) int {
	if month < 1 || month > 12 {
		return 0
	}
	if month == 2 {
		if IsLeapYear(year) {
			return 29
		}
		return 28
	}
	if month == 4 || month == 6 || month == 9 || month == 11 {
		return 30
	}
	return 31
}

// StartOfDay returns the start of the day for the given time
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of the day for the given time
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfWeek returns the start of the week for the given time
// The week is considered to start on Sunday by default
func StartOfWeek(t time.Time, startDay time.Weekday) time.Time {
	weekday := t.Weekday()
	if weekday == startDay {
		return StartOfDay(t)
	}
	daysToSubtract := (7 + int(weekday) - int(startDay)) % 7
	return StartOfDay(t.AddDate(0, 0, -daysToSubtract))
}

// EndOfWeek returns the end of the week for the given time
// The week is considered to end on Saturday by default
func EndOfWeek(t time.Time, endDay time.Weekday) time.Time {
	weekday := t.Weekday()
	if weekday == endDay {
		return EndOfDay(t)
	}
	daysToAdd := (7 + int(endDay) - int(weekday)) % 7
	return EndOfDay(t.AddDate(0, 0, daysToAdd))
}

// StartOfMonth returns the start of the month for the given time
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns the end of the month for the given time
func EndOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, -1).Add(time.Hour*23 + time.Minute*59 + time.Second*59 + time.Nanosecond*999999999)
}

// StartOfYear returns the start of the year for the given time
func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear returns the end of the year for the given time
func EndOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 12, 31, 23, 59, 59, 999999999, t.Location())
}

// FormatDuration formats a duration into a human-readable string
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return d.String()
	}

	parts := []string{}
	if d >= Year {
		years := d / Year
		parts = append(parts, fmt.Sprintf("%d year%s", years, pluralS(years)))
		d -= years * Year
	}
	if d >= Month {
		months := d / Month
		parts = append(parts, fmt.Sprintf("%d month%s", months, pluralS(months)))
		d -= months * Month
	}
	if d >= Week {
		weeks := d / Week
		parts = append(parts, fmt.Sprintf("%d week%s", weeks, pluralS(weeks)))
		d -= weeks * Week
	}
	if d >= Day {
		days := d / Day
		parts = append(parts, fmt.Sprintf("%d day%s", days, pluralS(days)))
		d -= days * Day
	}
	if d >= Hour {
		hours := d / Hour
		parts = append(parts, fmt.Sprintf("%d hour%s", hours, pluralS(hours)))
		d -= hours * Hour
	}
	if d >= Minute {
		minutes := d / Minute
		parts = append(parts, fmt.Sprintf("%d minute%s", minutes, pluralS(minutes)))
		d -= minutes * Minute
	}
	if d >= Second {
		seconds := d / Second
		parts = append(parts, fmt.Sprintf("%d second%s", seconds, pluralS(seconds)))
	}

	return strings.Join(parts, " ")
}

// pluralS returns "s" if the count is not 1
func pluralS(count time.Duration) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// ParseDuration parses a duration string and returns the time.Duration
// It supports years (y), months (M), weeks (w), days (d), hours (h), minutes (m), and seconds (s)
func ParseDuration(s string) (time.Duration, error) {
	var d time.Duration
	var err error

	for s != "" {
		var v int64
		var unit string

		if v, s, err = nextNumber(s); err != nil {
			return 0, err
		}
		if unit, s = nextUnit(s); unit == "" {
			return 0, fmt.Errorf("missing unit in duration %q", s)
		}

		switch unit {
		case "y":
			d += time.Duration(v) * Year
		case "M":
			d += time.Duration(v) * Month
		case "w":
			d += time.Duration(v) * Week
		case "d":
			d += time.Duration(v) * Day
		case "h":
			d += time.Duration(v) * time.Hour
		case "m":
			d += time.Duration(v) * time.Minute
		case "s":
			d += time.Duration(v) * time.Second
		default:
			return 0, fmt.Errorf("unknown unit %q in duration %q", unit, s)
		}
	}

	return d, nil
}

func nextNumber(s string) (int64, string, error) {
	i := 0
	for ; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			break
		}
	}
	if i == 0 {
		return 0, s, errors.New("invalid syntax")
	}
	n, err := strconv.ParseInt(s[:i], 10, 64)
	if err != nil {
		return 0, s, err
	}
	return n, s[i:], nil
}

func nextUnit(s string) (string, string) {
	if len(s) == 0 {
		return "", ""
	}
	return s[:1], s[1:]
}

// AddDate adds the specified number of years, months, and days to the given time
func AddDate(t time.Time, years, months, days int) time.Time {
	return t.AddDate(years, months, days)
}

// DaysBetween calculates the number of days between two dates
func DaysBetween(a, b time.Time) int {
	if a.After(b) {
		a, b = b, a
	}
	days := 0
	for a.Before(b) {
		a = a.AddDate(0, 0, 1)
		days++
	}
	return days
}

// IsSameDay checks if two times are on the same day
func IsSameDay(a, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// TimeIn returns the time in the specified timezone
func TimeIn(t time.Time, name string) (time.Time, error) {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return t, err
	}
	return t.In(loc), nil
}
