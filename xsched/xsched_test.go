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

	var mu sync.Mutex
	executed := false
	executedChan := make(chan struct{})

	id, err := c.AddFunc("* * * * * *", func() {
		mu.Lock()
		defer mu.Unlock()
		if !executed {
			executed = true
			close(executedChan)
		}
	})
	require.NoError(t, err)

	c.Start()

	select {
	case <-executedChan:
		// Job executed successfully
	case <-time.After(1100 * time.Millisecond):
		t.Fatal("Job should have been executed within the timeout period")
	}

	c.Stop()

	mu.Lock()
	assert.True(t, executed, "Job should have been executed with custom tick interval")
	mu.Unlock()

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
	executionDuration := 1100 * time.Millisecond // Increased duration to allow job execution

	wg.Add(jobCount)
	for i := 0; i < jobCount; i++ {
		_, err := c.AddFunc("* * * * * *", func() {
			mu.Lock()
			count++
			mu.Unlock()
			wg.Done()
		})
		require.NoError(t, err)
	}

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
	c := NewWithTickInterval(50 * time.Millisecond)
	var mu sync.Mutex
	jobExecutionCount := make(map[int]int)

	// Define expected executions for each job
	expectedExecutions := map[int]int{
		1: 4, // every 1 second
		2: 2, // every 2 seconds
		3: 1, // every 3 seconds
	}

	// Calculate total executions needed
	totalExecutions := 0
	for _, count := range expectedExecutions {
		totalExecutions += count
	}

	var wg sync.WaitGroup
	wg.Add(totalExecutions)

	allDone := make(chan struct{})

	// Add jobs with different schedules
	for jobID, expectedCount := range expectedExecutions {
		jobID := jobID
		expectedCount := expectedCount
		_, err := c.AddFunc(fmt.Sprintf("*/%d * * * * *", jobID), func() {
			mu.Lock()
			jobExecutionCount[jobID]++
			currentCount := jobExecutionCount[jobID]
			mu.Unlock()

			if currentCount <= expectedCount {
				wg.Done()
			}
		})
		require.NoError(t, err)
	}

	c.Start()

	go func() {
		wg.Wait()
		close(allDone)
	}()

	// Wait for all expected executions or timeout
	select {
	case <-allDone:
		// All expected executions completed
	case <-time.After(10 * time.Second):
		t.Fatal("Test timed out waiting for job executions")
	}

	c.Stop()

	// Verify each job executed as expected
	for jobID, expected := range expectedExecutions {
		mu.Lock()
		count := jobExecutionCount[jobID]
		mu.Unlock()
		assert.GreaterOrEqual(t, count, expected, fmt.Sprintf("Job %d should have executed at least %d times, got %d", jobID, expected, count))
	}
}

func TestCronStartStop(t *testing.T) {
	c := NewWithTickInterval(50 * time.Millisecond)

	executed := make(chan bool, 1)
	_, err := c.AddFunc("* * * * * *", func() {
		executed <- true
	})
	require.NoError(t, err)

	// Start the cron
	c.Start()
	assert.True(t, c.running, "Cron should be running after Start")

	// Wait for the job to execute
	select {
	case <-executed:
		// Job executed successfully
	case <-time.After(1100 * time.Millisecond): // Increased timeout to allow for job execution
		t.Fatal("Job should have been executed")
	}

	// Stop the cron
	c.Stop()
	assert.False(t, c.running, "Cron should not be running after Stop")

	// Ensure no further executions occur after stopping
	select {
	case <-executed:
		t.Fatal("Job should not execute after Stop")
	case <-time.After(200 * time.Millisecond):
		// No execution as expected
	}
}

func TestCronBuilder(t *testing.T) {
	t.Run("basic builder operations", func(t *testing.T) {
		schedule, err := NewCronBuilder().
			WithSeconds(0).
			WithMinutes(30).
			WithHours(12).
			WithDaysOfMonth(1).
			WithMonths(1, 6, 12).
			WithDaysOfWeek(1, 3, 5).
			Build()

		assert.NoError(t, err)
		assert.NotNil(t, schedule)

		cronSchedule, ok := schedule.(*cronSchedule)
		assert.True(t, ok)
		assert.Equal(t, []int{0}, cronSchedule.second)
		assert.Equal(t, []int{30}, cronSchedule.minute)
		assert.Equal(t, []int{12}, cronSchedule.hour)
		assert.Equal(t, []int{1}, cronSchedule.dayOfMonth)
		assert.Equal(t, []int{1, 6, 12}, cronSchedule.month)
		assert.Equal(t, []int{1, 3, 5}, cronSchedule.dayOfWeek)
	})

	t.Run("convenience methods", func(t *testing.T) {
		testCases := []struct {
			name     string
			builder  func() *CronBuilder
			validate func(*testing.T, Schedule)
		}{
			{
				name: "WithEverySecond",
				builder: func() *CronBuilder {
					return NewCronBuilder().WithEverySecond()
				},
				validate: func(t *testing.T, s Schedule) {
					cs := s.(*cronSchedule)
					assert.Len(t, cs.second, 60)
				},
			},
			{
				name: "WithEveryMinute",
				builder: func() *CronBuilder {
					return NewCronBuilder().WithEveryMinute()
				},
				validate: func(t *testing.T, s Schedule) {
					cs := s.(*cronSchedule)
					assert.Equal(t, []int{0}, cs.second)
					assert.Len(t, cs.minute, 60)
				},
			},
			{
				name: "WithEveryHour",
				builder: func() *CronBuilder {
					return NewCronBuilder().WithEveryHour()
				},
				validate: func(t *testing.T, s Schedule) {
					cs := s.(*cronSchedule)
					assert.Equal(t, []int{0}, cs.second)
					assert.Equal(t, []int{0}, cs.minute)
					assert.Len(t, cs.hour, 24)
				},
			},
			{
				name: "WithEveryDay",
				builder: func() *CronBuilder {
					return NewCronBuilder().WithEveryDay()
				},
				validate: func(t *testing.T, s Schedule) {
					cs := s.(*cronSchedule)
					assert.Equal(t, []int{0}, cs.second)
					assert.Equal(t, []int{0}, cs.minute)
					assert.Equal(t, []int{0}, cs.hour)
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				schedule, err := tc.builder().Build()
				assert.NoError(t, err)
				assert.NotNil(t, schedule)
				tc.validate(t, schedule)
			})
		}
	})

	t.Run("interval operations", func(t *testing.T) {
		schedule, err := NewCronBuilder().
			WithInterval("second", 0, 15).
			WithInterval("minute", 0, 30).
			WithInterval("hour", 0, 4).
			Build()

		assert.NoError(t, err)
		assert.NotNil(t, schedule)

		cs := schedule.(*cronSchedule)
		assert.Equal(t, []int{0, 15, 30, 45}, cs.second)
		assert.Equal(t, []int{0, 30}, cs.minute)
		assert.Equal(t, []int{0, 4, 8, 12, 16, 20}, cs.hour)
	})

	t.Run("error handling", func(t *testing.T) {
		testCases := []struct {
			name    string
			builder func() (*CronBuilder, error)
		}{
			{
				name: "invalid seconds",
				builder: func() (*CronBuilder, error) {
					return NewCronBuilder().WithSeconds(60), nil
				},
			},
			{
				name: "invalid minutes",
				builder: func() (*CronBuilder, error) {
					return NewCronBuilder().WithMinutes(-1), nil
				},
			},
			{
				name: "invalid hours",
				builder: func() (*CronBuilder, error) {
					return NewCronBuilder().WithHours(24), nil
				},
			},
			{
				name: "invalid interval field",
				builder: func() (*CronBuilder, error) {
					return NewCronBuilder().WithInterval("invalid", 0, 1), nil
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder, _ := tc.builder()
				_, err := builder.Build()
				assert.Error(t, err)
			})
		}
	})
}

func TestCronExpression(t *testing.T) {
	t.Run("parse expression", func(t *testing.T) {
		expr, err := ParseExpression("*/15 0 */4 * * 1-5")
		assert.NoError(t, err)
		assert.NotNil(t, expr)

		assert.Equal(t, []int{0, 15, 30, 45}, expr.Second.Valid)
		assert.Equal(t, []int{0}, expr.Minute.Valid)
		assert.Equal(t, []int{0, 4, 8, 12, 16, 20}, expr.Hour.Valid)
		assert.Equal(t, []int{1, 2, 3, 4, 5}, expr.DayOfWeek.Valid)
	})

	t.Run("validate expression", func(t *testing.T) {
		testCases := []struct {
			spec  string
			valid bool
		}{
			{"* * * * * *", true},
			{"*/15 * * * * *", true},
			{"0 0 0 1 * *", true},
			{"invalid", false},
			{"* * * *", false},
			{"60 * * * * *", false},
		}

		for _, tc := range testCases {
			t.Run(tc.spec, func(t *testing.T) {
				err := ValidateExpression(tc.spec)
				if tc.valid {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})

	t.Run("is valid time", func(t *testing.T) {
		expr, err := ParseExpression("0 30 9 * * 1-5")
		assert.NoError(t, err)

		validTime := time.Date(2023, 5, 1, 9, 30, 0, 0, time.UTC)   // Monday 9:30:00
		invalidTime := time.Date(2023, 5, 6, 9, 30, 0, 0, time.UTC) // Saturday 9:30:00

		assert.True(t, expr.IsValid(validTime))
		assert.False(t, expr.IsValid(invalidTime))
	})

	t.Run("build schedule from spec", func(t *testing.T) {
		schedule, err := BuildScheduleFromSpec("*/15 * * * * *")
		assert.NoError(t, err)
		assert.NotNil(t, schedule)

		now := time.Now()
		next := schedule.Next(now)
		assert.True(t, next.After(now))
	})
}

func TestCronBuilderIntegration(t *testing.T) {
	c := New()
	schedule, err := NewCronBuilder().
		WithSeconds(0).
		WithMinutes(0).
		WithHours(12).
		WithDaysOfWeek(1).
		Build()

	assert.NoError(t, err)

	var executed bool
	var mu sync.Mutex
	job := func() {
		mu.Lock()
		executed = true
		mu.Unlock()
	}

	c.addJobWithOptions(schedule, job, false)
	c.Start()

	// Wait for potential execution
	time.Sleep(100 * time.Millisecond)
	c.Stop()

	mu.Lock()
	wasExecuted := executed
	mu.Unlock()

	// Check if job executed at the right time
	now := time.Now()
	shouldExecute := now.Hour() == 12 &&
		now.Minute() == 0 &&
		now.Second() == 0 &&
		now.Weekday() == time.Monday

	assert.Equal(t, shouldExecute, wasExecuted)
}
