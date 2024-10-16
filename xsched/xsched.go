package xsched

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Job represents a function to be executed on a schedule
type Job func()

// Schedule represents when a job should be executed
type Schedule interface {
	Next(time.Time) time.Time
}

// Cron manages scheduled jobs
type Cron struct {
	jobs         []*jobEntry
	running      bool
	mutex        sync.Mutex
	location     *time.Location
	tickInterval time.Duration
}

// jobEntry represents a scheduled job
type jobEntry struct {
	schedule Schedule
	job      Job
	next     time.Time
	id       string
}

// New creates a new Cron instance with default tick interval of 1 second
func New() *Cron {
	return &Cron{
		jobs:         make([]*jobEntry, 0),
		location:     time.Local,
		tickInterval: time.Second,
	}
}

// NewWithTickInterval creates a new Cron instance with a custom tick interval
func NewWithTickInterval(interval time.Duration) *Cron {
	return &Cron{
		jobs:         make([]*jobEntry, 0),
		location:     time.Local,
		tickInterval: interval,
	}
}

// AddFunc adds a function to be executed on the given schedule
func (c *Cron) AddFunc(spec string, cmd func()) (string, error) {
	schedule, err := parseSchedule(spec)
	if err != nil {
		return "", err
	}
	return c.addJob(schedule, Job(cmd)), nil
}

// addJob adds a job to be run on the given schedule
func (c *Cron) addJob(schedule Schedule, cmd Job) string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	job := &jobEntry{
		schedule: schedule,
		job:      cmd,
		id:       fmt.Sprintf("%p", cmd),
	}

	now := time.Now().In(c.location)
	job.next = schedule.Next(now) // Initialize next execution time
	c.jobs = append(c.jobs, job)

	return job.id
}

// Remove removes a job by its ID
func (c *Cron) Remove(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i, job := range c.jobs {
		if job.id == id {
			c.jobs = append(c.jobs[:i], c.jobs[i+1:]...)
			break
		}
	}
}

// Start begins the cron scheduler
func (c *Cron) Start() {
	c.mutex.Lock()
	if c.running {
		c.mutex.Unlock()
		return
	}
	c.running = true
	c.mutex.Unlock()
	go c.run()
}

// run executes the main scheduling loop
func (c *Cron) run() {
	for {
		c.mutex.Lock()
		if !c.running {
			c.mutex.Unlock()
			break
		}

		now := time.Now().In(c.location)
		var duration time.Duration

		if len(c.jobs) == 0 {
			c.mutex.Unlock()
			time.Sleep(c.tickInterval)
			continue
		}

		// Find the next job to run
		earliestTime := time.Time{}
		for _, job := range c.jobs {
			if earliestTime.IsZero() || job.next.Before(earliestTime) {
				earliestTime = job.next
			}
		}

		if earliestTime.After(now) {
			duration = earliestTime.Sub(now)
		} else {
			duration = 0
		}

		c.mutex.Unlock()

		if duration > 0 {
			time.Sleep(duration)
		} else {
			// Run due jobs
			c.mutex.Lock()
			now = time.Now().In(c.location)
			for _, job := range c.jobs {
				if !job.next.After(now) {
					go job.job()
					// Update next execution time based on the current job's schedule
					job.next = job.schedule.Next(job.next)
				}
			}
			c.mutex.Unlock()
		}
	}
}

// Stop halts the cron scheduler
func (c *Cron) Stop() {
	c.mutex.Lock()
	c.running = false
	c.mutex.Unlock()
}

// parseSchedule parses a cron schedule specification
func parseSchedule(spec string) (Schedule, error) {
	fields := strings.Fields(spec)
	if len(fields) != 6 {
		return nil, fmt.Errorf("invalid cron spec, expected 6 fields, got %d", len(fields))
	}

	schedule := &cronSchedule{}

	var err error
	schedule.second, err = parseField(fields[0], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("invalid second field: %v", err)
	}
	schedule.minute, err = parseField(fields[1], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("invalid minute field: %v", err)
	}
	schedule.hour, err = parseField(fields[2], 0, 23)
	if err != nil {
		return nil, fmt.Errorf("invalid hour field: %v", err)
	}
	schedule.dayOfMonth, err = parseField(fields[3], 1, 31)
	if err != nil {
		return nil, fmt.Errorf("invalid day of month field: %v", err)
	}
	schedule.month, err = parseField(fields[4], 1, 12)
	if err != nil {
		return nil, fmt.Errorf("invalid month field: %v", err)
	}
	schedule.dayOfWeek, err = parseField(fields[5], 0, 6)
	if err != nil {
		return nil, fmt.Errorf("invalid day of week field: %v", err)
	}

	return schedule, nil
}

// cronSchedule implements the Schedule interface
type cronSchedule struct {
	second, minute, hour, dayOfMonth, month, dayOfWeek []int
}

// Next returns the next activation time, later than the given time
func (s *cronSchedule) Next(t time.Time) time.Time {
	loc := t.Location()
	t = t.Truncate(time.Second).Add(time.Second) // Round up to the next whole second

	// Set an upper limit for searching to prevent infinite loops
	endTime := t.Add(5 * 365 * 24 * time.Hour) // 5 years into the future

	for t.Before(endTime) {
		month, day := t.Month(), t.Day()
		hour, min, sec := t.Clock()

		// Check if the current time satisfies the schedule
		if s.matchField(s.month, int(month)) &&
			s.matchField(s.dayOfMonth, day) &&
			s.matchField(s.dayOfWeek, int(t.Weekday())) &&
			s.matchField(s.hour, hour) &&
			s.matchField(s.minute, min) &&
			s.matchField(s.second, sec) {
			return t.In(loc)
		}

		// Increment by one second
		t = t.Add(time.Second)
	}

	// If no matching time is found within the limit
	return time.Time{}
}

// matchField checks if the value matches the field values
func (s *cronSchedule) matchField(field []int, value int) bool {
	if len(field) == 0 {
		return true
	}
	for _, v := range field {
		if v == value {
			return true
		}
	}
	return false
}

// parseField parses a cron field into a slice of integers
func parseField(field string, min int, max int) ([]int, error) {
	var result []int
	if field == "*" {
		for i := min; i <= max; i++ {
			result = append(result, i)
		}
		return result, nil
	} else if strings.Contains(field, "/") {
		parts := strings.Split(field, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid step expression in field: %s", field)
		}
		step, err := strconv.Atoi(parts[1])
		if err != nil || step <= 0 {
			return nil, fmt.Errorf("invalid step value in field: %s", field)
		}
		var rangeStart, rangeEnd int
		if parts[0] == "*" {
			rangeStart = min
			rangeEnd = max
		} else {
			rangeParts := strings.Split(parts[0], "-")
			if len(rangeParts) == 2 {
				rangeStart, err = strconv.Atoi(rangeParts[0])
				if err != nil {
					return nil, fmt.Errorf("invalid range start in field: %s", field)
				}
				rangeEnd, err = strconv.Atoi(rangeParts[1])
				if err != nil {
					return nil, fmt.Errorf("invalid range end in field: %s", field)
				}
			} else {
				rangeStart, err = strconv.Atoi(parts[0])
				if err != nil {
					return nil, fmt.Errorf("invalid range in field: %s", field)
				}
				rangeEnd = max
			}
		}
		if rangeStart < min || rangeEnd > max {
			return nil, fmt.Errorf("range outside valid bounds in field: %s", field)
		}
		for i := rangeStart; i <= rangeEnd; i += step {
			result = append(result, i)
		}
	} else if strings.Contains(field, "-") {
		rangeParts := strings.Split(field, "-")
		if len(rangeParts) != 2 {
			return nil, fmt.Errorf("invalid range expression in field: %s", field)
		}
		start, err := strconv.Atoi(rangeParts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid range start in field: %s", field)
		}
		end, err := strconv.Atoi(rangeParts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid range end in field: %s", field)
		}
		if start < min || end > max {
			return nil, fmt.Errorf("range outside valid bounds in field: %s", field)
		}
		if start > end {
			return nil, fmt.Errorf("start greater than end in field: %s", field)
		}
		for i := start; i <= end; i++ {
			result = append(result, i)
		}
	} else if strings.Contains(field, ",") {
		parts := strings.Split(field, ",")
		for _, part := range parts {
			val, err := strconv.Atoi(part)
			if err != nil || val < min || val > max {
				return nil, fmt.Errorf("invalid value in field: %s", field)
			}
			result = append(result, val)
		}
	} else {
		val, err := strconv.Atoi(field)
		if err != nil || val < min || val > max {
			return nil, fmt.Errorf("invalid value in field: %s", field)
		}
		result = append(result, val)
	}

	return result, nil
}

// AddEverySecond adds a job to be executed every second
func (c *Cron) AddEverySecond(cmd func()) (string, error) {
	return c.AddFunc("* * * * * *", cmd)
}

// AddEveryMinute adds a job to be executed every minute
func (c *Cron) AddEveryMinute(cmd func()) (string, error) {
	return c.AddFunc("0 * * * * *", cmd)
}

// AddEveryHour adds a job to be executed every hour
func (c *Cron) AddEveryHour(cmd func()) (string, error) {
	return c.AddFunc("0 0 * * * *", cmd)
}

// AddEveryDay adds a job to be executed every day
func (c *Cron) AddEveryDay(cmd func()) (string, error) {
	return c.AddFunc("0 0 0 * * *", cmd)
}

// AddEveryWeek adds a job to be executed every week (on Sunday)
func (c *Cron) AddEveryWeek(cmd func()) (string, error) {
	return c.AddFunc("0 0 0 * * 0", cmd)
}

// AddEveryMonth adds a job to be executed every month
func (c *Cron) AddEveryMonth(cmd func()) (string, error) {
	return c.AddFunc("0 0 0 1 * *", cmd)
}

// AddEveryNSeconds adds a job to be executed every N seconds
func (c *Cron) AddEveryNSeconds(n int, cmd func()) (string, error) {
	if n <= 0 || n > 59 {
		return "", fmt.Errorf("invalid interval: %d", n)
	}
	spec := fmt.Sprintf("*/%d * * * * *", n)
	return c.AddFunc(spec, cmd)
}
