package xsched

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCron(t *testing.T) {
	c := NewWithTickInterval(100 * time.Millisecond)

	// Test AddFunc with valid spec
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

func TestCronWithCustomTickInterval(t *testing.T) {
	c := NewWithTickInterval(50 * time.Millisecond)
	assert.Equal(t, 50*time.Millisecond, c.tickInterval, "Custom tick interval should be set correctly")

	executed := false
	id, err := c.AddFunc("* * * * * *", func() {
		executed = true
	})
	require.NoError(t, err)

	c.Start()
	time.Sleep(100 * time.Millisecond)
	c.Stop()

	assert.True(t, executed, "Job should have been executed with custom tick interval")
	c.Remove(id)
}

func TestRemoveJob(t *testing.T) {
	c := New()
	id1, _ := c.AddFunc("* * * * * *", func() {})
	id2, _ := c.AddFunc("*/2 * * * * *", func() {})

	assert.Len(t, c.jobs, 2, "Two jobs should be added")

	c.Remove(id1)
	assert.Len(t, c.jobs, 1, "One job should be removed")
	assert.Equal(t, id2, c.jobs[0].id, "Correct job should remain")

	c.Remove(id2)
	assert.Empty(t, c.jobs, "All jobs should be removed")
}

func TestConcurrentJobExecution(t *testing.T) {
	c := NewWithTickInterval(50 * time.Millisecond)
	var wg sync.WaitGroup
	var mu sync.Mutex
	count := 0

	jobCount := 5
	executionDuration := 300 * time.Millisecond

	for i := 0; i < jobCount; i++ {
		_, err := c.AddFunc("* * * * * *", func() {
			mu.Lock()
			count++
			mu.Unlock()
			wg.Done()
		})
		require.NoError(t, err)
	}

	wg.Add(jobCount)
	c.Start()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All jobs executed successfully
	case <-time.After(executionDuration):
		t.Fatal("Test timed out")
	}

	c.Stop()

	assert.GreaterOrEqual(t, count, jobCount, "All jobs should have executed at least once")
}

func TestNextExecutionEdgeCases(t *testing.T) {
	now := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)

	testCases := []struct {
		spec     string
		expected time.Time
	}{
		{"0 0 0 1 1 *", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"59 59 23 31 12 *", time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)},
		{"0 0 0 29 2 *", time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)}, // Leap year
		{"0 0 0 * * 1", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},   // Next Monday
	}

	for _, tc := range testCases {
		t.Run(tc.spec, func(t *testing.T) {
			schedule, err := parseSchedule(tc.spec)
			require.NoError(t, err)
			next := schedule.Next(now)
			assert.Equal(t, tc.expected, next, "Next execution time should be correct for spec: %s", tc.spec)
		})
	}
}

func TestInvalidScheduleSpecs(t *testing.T) {
	invalidSpecs := []string{
		"* * * *",       // Too few fields
		"* * * * * * *", // Too many fields
		"60 * * * * *",  // Invalid second
		"* 60 * * * *",  // Invalid minute
		"* * 24 * * *",  // Invalid hour
		"* * * 32 * *",  // Invalid day of month
		"* * * * 13 *",  // Invalid month
		"* * * * * 7",   // Invalid day of week
		"a * * * * *",   // Non-numeric value
		"*/0 * * * * *", // Invalid step value
	}

	for _, spec := range invalidSpecs {
		_, err := parseSchedule(spec)
		assert.Error(t, err, "Invalid spec '%s' should fail to parse", spec)
	}
}

func TestJobExecutionOrder(t *testing.T) {
	c := NewWithTickInterval(100 * time.Millisecond)
	var mu sync.Mutex
	var executionOrder []int

	for i := 1; i <= 3; i++ {
		i := i // Capture loop variable
		_, err := c.AddFunc(fmt.Sprintf("*/%d * * * * *", i), func() {
			mu.Lock()
			executionOrder = append(executionOrder, i)
			mu.Unlock()
		})
		require.NoError(t, err)
	}

	c.Start()
	time.Sleep(1100 * time.Millisecond) // Run for just over 1 second
	c.Stop()

	expected := []int{1, 2, 3, 1}
	assert.Equal(t, expected, executionOrder, "Jobs should execute in the correct order")
}

func TestCronStartStop(t *testing.T) {
	c := NewWithTickInterval(50 * time.Millisecond)

	executed := false
	_, err := c.AddFunc("* * * * * *", func() {
		executed = true
	})
	require.NoError(t, err)

	// Start the cron
	c.Start()
	assert.True(t, c.running, "Cron should be running after Start")

	// Wait for the job to execute
	time.Sleep(100 * time.Millisecond)

	// Stop the cron
	c.Stop()
	assert.False(t, c.running, "Cron should not be running after Stop")

	// Check if the job was executed
	assert.True(t, executed, "Job should have been executed")

	// Reset the executed flag
	executed = false

	// Wait to ensure the job doesn't execute after stopping
	time.Sleep(100 * time.Millisecond)
	assert.False(t, executed, "Job should not execute after Stop")
}
