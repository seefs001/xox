package xtime

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/seefs001/xox/xerror"
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
	if month < 1 || month > 12 || year < 1 {
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
	if t.IsZero() {
		return t
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of the day for the given time
func EndOfDay(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfWeek returns the start of the week for the given time
// The week is considered to start on Sunday by default
func StartOfWeek(t time.Time, startDay time.Weekday) time.Time {
	if t.IsZero() {
		return t
	}
	if startDay < time.Sunday || startDay > time.Saturday {
		startDay = time.Sunday
	}
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
	if t.IsZero() {
		return t
	}
	if endDay < time.Sunday || endDay > time.Saturday {
		endDay = time.Saturday
	}
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
	if s == "" {
		return 0, xerror.New("empty duration string")
	}

	var d time.Duration
	s = strings.TrimSpace(s)

	for s != "" {
		var v int64
		var unit string
		var err error

		if v, s, err = nextNumber(s); err != nil {
			return 0, xerror.Wrap(err, "failed to parse number")
		}
		if unit, s = nextUnit(s); unit == "" {
			return 0, xerror.New("missing unit in duration")
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
			d += time.Duration(v) * Hour
		case "m":
			d += time.Duration(v) * Minute
		case "s":
			d += time.Duration(v) * Second
		default:
			return 0, xerror.Newf("invalid unit %q in duration", unit)
		}
	}

	return d, nil
}

// nextNumber extracts the next number from the duration string
func nextNumber(s string) (int64, string, error) {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	s = s[i:]
	if len(s) == 0 {
		return 0, "", xerror.New("empty number")
	}

	i = 0
	for i < len(s) && (s[i] >= '0' && s[i] <= '9') {
		i++
	}
	if i == 0 {
		return 0, "", xerror.New("invalid number format")
	}

	n, err := strconv.ParseInt(s[:i], 10, 64)
	if err != nil {
		return 0, "", xerror.Wrap(err, "failed to parse number")
	}
	if n < 0 {
		return 0, "", xerror.New("negative duration not allowed")
	}

	return n, s[i:], nil
}

// nextUnit extracts the next unit from the duration string
func nextUnit(s string) (string, string) {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	s = s[i:]
	if len(s) == 0 {
		return "", ""
	}

	switch s[0] {
	case 'y', 'M', 'w', 'd', 'h', 'm', 's':
		return string(s[0]), s[1:]
	default:
		return "", s
	}
}

// AddDate adds the specified number of years, months, and days to the given time
func AddDate(t time.Time, years, months, days int) time.Time {
	if t.IsZero() {
		return t
	}
	return t.AddDate(years, months, days)
}

// DaysBetween calculates the number of days between two dates
// The result is always positive, regardless of the order of dates
func DaysBetween(a, b time.Time) int {
	if a.IsZero() || b.IsZero() {
		return 0
	}

	// Normalize both times to start of day in UTC to ensure accurate day calculation
	aDay := StartOfDay(a).UTC()
	bDay := StartOfDay(b).UTC()

	days := int(bDay.Sub(aDay).Hours() / 24)
	if days < 0 {
		days = -days
	}
	return days
}

// IsSameDay checks if two times are on the same day
func IsSameDay(a, b time.Time) bool {
	if a.IsZero() || b.IsZero() {
		return false
	}
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

// TimeIn returns the time in the specified timezone
func TimeIn(t time.Time, name string) (time.Time, error) {
	if t.IsZero() {
		return t, nil
	}

	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.Time{}, xerror.Wrapf(err, "failed to load timezone %q", name)
	}
	return t.In(loc), nil
}

// IsWeekend checks if the given time falls on a weekend (Saturday or Sunday)
func IsWeekend(t time.Time) bool {
	if t.IsZero() {
		return false
	}
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// Quarter returns the quarter (1-4) for the given time
func Quarter(t time.Time) int {
	if t.IsZero() {
		return 0
	}
	return int(t.Month()-1)/3 + 1
}

// StartOfQuarter returns the start of the quarter for the given time
func StartOfQuarter(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	quarter := Quarter(t)
	month := time.Month((quarter-1)*3 + 1)
	return time.Date(t.Year(), month, 1, 0, 0, 0, 0, t.Location())
}

// EndOfQuarter returns the end of the quarter for the given time
func EndOfQuarter(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return StartOfQuarter(t).AddDate(0, 3, 0).Add(-time.Nanosecond)
}
