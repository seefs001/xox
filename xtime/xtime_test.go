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

	// Check if the instant in time remains the same
	assert.Equal(t, utcTime.Unix(), nycTime.Unix())

	// Check if the hour is correct for New York time
	expectedHour := 19 // 7 PM on the previous day in New York
	assert.Equal(t, expectedHour, nycTime.Hour())

	// Check if the date and time are correct for New York time
	expectedDateTime := time.Date(2022, 12, 31, 19, 0, 0, 0, nycTime.Location())
	assert.True(t, nycTime.Equal(expectedDateTime), "Expected %v, but got %v", expectedDateTime, nycTime)

	// Additional check to print actual times for debugging
	t.Logf("UTC time: %v, NYC time: %v", utcTime, nycTime)

	// Test invalid timezone
	_, err = xtime.TimeIn(utcTime, "Invalid/Timezone")
	assert.Error(t, err)
}

func TestStartOfWeek(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		startDay time.Weekday
		expected time.Time
	}{
		{
			name:     "Sunday start",
			input:    time.Date(2023, 5, 10, 12, 0, 0, 0, time.UTC), // Wednesday
			startDay: time.Sunday,
			expected: time.Date(2023, 5, 7, 0, 0, 0, 0, time.UTC), // Previous Sunday
		},
		{
			name:     "Monday start",
			input:    time.Date(2023, 5, 10, 12, 0, 0, 0, time.UTC), // Wednesday
			startDay: time.Monday,
			expected: time.Date(2023, 5, 8, 0, 0, 0, 0, time.UTC), // Previous Monday
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := xtime.StartOfWeek(tt.input, tt.startDay)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEndOfWeek(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		endDay   time.Weekday
		expected time.Time
	}{
		{
			name:     "Saturday end",
			input:    time.Date(2023, 5, 10, 12, 0, 0, 0, time.UTC), // Wednesday
			endDay:   time.Saturday,
			expected: time.Date(2023, 5, 13, 23, 59, 59, 999999999, time.UTC), // Next Saturday
		},
		{
			name:     "Sunday end",
			input:    time.Date(2023, 5, 10, 12, 0, 0, 0, time.UTC), // Wednesday
			endDay:   time.Sunday,
			expected: time.Date(2023, 5, 14, 23, 59, 59, 999999999, time.UTC), // Next Sunday
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := xtime.EndOfWeek(tt.input, tt.endDay)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStartOfMonth(t *testing.T) {
	input := time.Date(2023, 5, 15, 12, 30, 0, 0, time.UTC)
	expected := time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC)
	result := xtime.StartOfMonth(input)
	assert.Equal(t, expected, result)
}

func TestEndOfMonth(t *testing.T) {
	input := time.Date(2023, 5, 15, 12, 30, 0, 0, time.UTC)
	expected := time.Date(2023, 5, 31, 23, 59, 59, 999999999, time.UTC)
	result := xtime.EndOfMonth(input)
	assert.Equal(t, expected, result)
}

func TestStartOfYear(t *testing.T) {
	input := time.Date(2023, 5, 15, 12, 30, 0, 0, time.UTC)
	expected := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	result := xtime.StartOfYear(input)
	assert.Equal(t, expected, result)
}

func TestEndOfYear(t *testing.T) {
	input := time.Date(2023, 5, 15, 12, 30, 0, 0, time.UTC)
	expected := time.Date(2023, 12, 31, 23, 59, 59, 999999999, time.UTC)
	result := xtime.EndOfYear(input)
	assert.Equal(t, expected, result)
}

func TestAddDate(t *testing.T) {
	input := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	expected := time.Date(2024, 3, 11, 0, 0, 0, 0, time.UTC)
	result := xtime.AddDate(input, 1, 2, 10)
	assert.Equal(t, expected, result)
}

func TestTimeIn_InvalidTimezone(t *testing.T) {
	utcTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := xtime.TimeIn(utcTime, "Invalid/Timezone")
	assert.Error(t, err)
}
