package xtime_test

import (
	"testing"
	"time"

	"github.com/seefs001/xox/xtime"
	"github.com/stretchr/testify/assert"
)

func TestIsLeapYear(t *testing.T) {
	tests := []struct {
		year     int
		expected bool
	}{
		{2000, true},
		{2001, false},
		{2004, true},
		{2100, false},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, xtime.IsLeapYear(test.year), "IsLeapYear(%d)", test.year)
	}
}

func TestDaysInMonth(t *testing.T) {
	tests := []struct {
		year     int
		month    int
		expected int
	}{
		{2023, 1, 31},
		{2023, 2, 28},
		{2024, 2, 29},
		{2023, 4, 30},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, xtime.DaysInMonth(test.year, test.month), "DaysInMonth(%d, %d)", test.year, test.month)
	}
}

func TestStartOfDay(t *testing.T) {
	now := time.Now()
	start := xtime.StartOfDay(now)
	assert.Equal(t, 0, start.Hour())
	assert.Equal(t, 0, start.Minute())
	assert.Equal(t, 0, start.Second())
	assert.Equal(t, 0, start.Nanosecond())
}

func TestEndOfDay(t *testing.T) {
	now := time.Now()
	end := xtime.EndOfDay(now)
	assert.Equal(t, 23, end.Hour())
	assert.Equal(t, 59, end.Minute())
	assert.Equal(t, 59, end.Second())
	assert.Equal(t, 999999999, end.Nanosecond())
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{time.Hour*24*365 + time.Hour*24*30 + time.Hour*24*7 + time.Hour*24 + time.Hour + time.Minute + time.Second, "1 year 1 month 1 week 1 day 1 hour 1 minute 1 second"},
		{time.Hour * 2, "2 hours"},
		{time.Minute * 30, "30 minutes"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, xtime.FormatDuration(test.duration))
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"1y2M3w4d5h6m7s", xtime.Year + 2*xtime.Month + 3*xtime.Week + 4*xtime.Day + 5*time.Hour + 6*time.Minute + 7*time.Second},
		{"24h", 24 * time.Hour},
		{"30m", 30 * time.Minute},
	}

	for _, test := range tests {
		result, err := xtime.ParseDuration(test.input)
		assert.NoError(t, err)
		assert.Equal(t, test.expected, result)
	}
}

func TestDaysBetween(t *testing.T) {
	date1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2023, 1, 10, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, 9, xtime.DaysBetween(date1, date2))
	assert.Equal(t, 9, xtime.DaysBetween(date2, date1)) // Should work in reverse order too
}

func TestIsSameDay(t *testing.T) {
	date1 := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	date2 := time.Date(2023, 1, 1, 22, 0, 0, 0, time.UTC)
	date3 := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	assert.True(t, xtime.IsSameDay(date1, date2))
	assert.False(t, xtime.IsSameDay(date1, date3))
}

func TestTimeIn(t *testing.T) {
	utcTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	nycTime, err := xtime.TimeIn(utcTime, "America/New_York")
	assert.NoError(t, err)

	// Check if the timezone has been correctly applied
	assert.Equal(t, "America/New_York", nycTime.Location().String())

	// FIXME: Calculate the expected offset
	// Calculate the expected offset
	// _, offset := nycTime.Zone()
	// expectedOffset := time.Duration(-offset) * time.Second

	// // Check if the time difference is as expected
	// assert.Equal(t, expectedOffset, nycTime.Sub(utcTime), "Time difference should be %v, but got %v", expectedOffset, nycTime.Sub(utcTime))

	// Additional check to print actual times for debugging
	t.Logf("UTC time: %v, NYC time: %v", utcTime, nycTime)

	_, err = xtime.TimeIn(utcTime, "Invalid/Timezone")
	assert.Error(t, err)
}
