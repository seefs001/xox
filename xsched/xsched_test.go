package xsched

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCron(t *testing.T) {
	c := New()

	// Test AddFunc
	id, err := c.AddFunc("* * * * * *", func() {})
	assert.NoError(t, err, "AddFunc should not fail with valid spec")
	assert.NotEmpty(t, id, "AddFunc should return a non-empty job ID")

	// Test invalid cron spec
	id, err = c.AddFunc("invalid spec", func() {})
	assert.Error(t, err, "AddFunc should fail with invalid spec")
	assert.Empty(t, id, "AddFunc should return an empty job ID for invalid spec")

	// Test Start and Stop
	c.Start()
	assert.True(t, c.running, "Cron should be running after Start")

	c.Stop()
	assert.False(t, c.running, "Cron should not be running after Stop")

	// Test job execution
	executed := false
	var mu sync.Mutex
	id, err = c.AddFunc("* * * * * *", func() {
		mu.Lock()
		executed = true
		mu.Unlock()
	})
	require.NoError(t, err)

	c.Start()

	// Use a channel to wait for job execution
	done := make(chan bool)
	go func() {
		for i := 0; i < 20; i++ { // Try for 2 seconds
			mu.Lock()
			if executed {
				mu.Unlock()
				done <- true
				return
			}
			mu.Unlock()
			time.Sleep(100 * time.Millisecond)
		}
		done <- false
	}()

	select {
	case result := <-done:
		assert.True(t, result, "Job should have been executed")
	case <-time.After(3 * time.Second):
		t.Fatal("Test timed out")
	}

	c.Stop()

	// Test Remove
	c = New()
	id, err = c.AddFunc("* * * * * *", func() {})
	require.NoError(t, err)

	assert.Len(t, c.jobs, 1, "Job should have been added")

	c.Remove(id)

	assert.Empty(t, c.jobs, "Job should have been removed")
}

func TestParseSchedule(t *testing.T) {
	testCases := []struct {
		spec    string
		isValid bool
	}{
		{"* * * * * *", true},
		{"0 0 0 1 1 *", true},
		{"*/15 * * * * *", true},
		{"0 0 0 * * 1-5", true},
		{"invalid", false},
		{"* * * * *", false},
		{"60 * * * * *", false},
	}

	for _, tc := range testCases {
		_, err := parseSchedule(tc.spec)
		if tc.isValid {
			assert.NoError(t, err, "Valid spec '%s' should parse without error", tc.spec)
		} else {
			assert.Error(t, err, "Invalid spec '%s' should fail to parse", tc.spec)
		}
	}
}

func TestNextExecution(t *testing.T) {
	c := New()
	_, err := c.AddFunc("0 0 0 * * *", func() {}) // Every day at midnight
	require.NoError(t, err)

	now := time.Date(2023, 5, 1, 12, 0, 0, 0, time.UTC)
	next := c.jobs[0].schedule.Next(now)

	expected := time.Date(2023, 5, 2, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, next, "Next execution time should be correct")
}

func TestConvenienceFunctions(t *testing.T) {
	c := New()

	testCases := []struct {
		name     string
		addFunc  func(func()) (string, error)
		expected string
	}{
		{"AddEverySecond", c.AddEverySecond, "* * * * * *"},
		{"AddEveryMinute", c.AddEveryMinute, "0 * * * * *"},
		{"AddEveryHour", c.AddEveryHour, "0 0 * * * *"},
		{"AddEveryDay", c.AddEveryDay, "0 0 0 * * *"},
		{"AddEveryWeek", c.AddEveryWeek, "0 0 0 * * 0"},
		{"AddEveryMonth", c.AddEveryMonth, "0 0 0 1 * *"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := tc.addFunc(func() {})
			assert.NoError(t, err)
			assert.NotEmpty(t, id)

			job := c.jobs[len(c.jobs)-1]
			schedule, ok := job.schedule.(*cronSchedule)
			assert.True(t, ok)

			expected, err := parseSchedule(tc.expected)
			assert.NoError(t, err)
			expectedCron, ok := expected.(*cronSchedule)
			assert.True(t, ok)

			assert.Equal(t, expectedCron, schedule)
		})
	}
}

func TestAddEveryNSeconds(t *testing.T) {
	c := New()
	id, err := c.AddEveryNSeconds(30, func() {})
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	job := c.jobs[len(c.jobs)-1]
	schedule, ok := job.schedule.(*cronSchedule)
	assert.True(t, ok)

	expected, err := parseSchedule("*/30 * * * * *")
	assert.NoError(t, err)
	expectedCron, ok := expected.(*cronSchedule)
	assert.True(t, ok)

	assert.Equal(t, expectedCron, schedule)
}
