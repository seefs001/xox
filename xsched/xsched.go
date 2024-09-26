package xsched

import (
	"fmt"
	"sort"
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
	jobs     []*jobEntry
	stop     chan struct{}
	add      chan *jobEntry
	remove   chan string
	running  bool
	mutex    sync.RWMutex
	location *time.Location
}

// jobEntry represents a scheduled job
type jobEntry struct {
	schedule Schedule
	job      Job
	next     time.Time
	id       string
}

// New creates a new Cron instance
func New() *Cron {
	return &Cron{
		jobs:     make([]*jobEntry, 0),
		add:      make(chan *jobEntry),
		stop:     make(chan struct{}),
		remove:   make(chan string),
		running:  false,
		location: time.Local,
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

	if !c.running {
		c.jobs = append(c.jobs, job)
	} else {
		c.add <- job
	}

	return job.id
}

// Remove removes a job by its ID
func (c *Cron) Remove(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.running {
		c.remove <- id
	} else {
		c.removeJob(id)
	}
}

// Start begins the cron scheduler
func (c *Cron) Start() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.running {
		return
	}
	c.running = true
	go c.run()
}

// run executes the main scheduling loop
func (c *Cron) run() {
	now := time.Now().In(c.location)
	c.mutex.Lock()
	for _, job := range c.jobs {
		job.next = job.schedule.Next(now)
	}
	c.mutex.Unlock()

	for {
		now = time.Now().In(c.location)
		var timer *time.Timer
		c.mutex.RLock()
		if len(c.jobs) == 0 {
			timer = time.NewTimer(time.Minute)
		} else {
			sort.Slice(c.jobs, func(i, j int) bool {
				return c.jobs[i].next.Before(c.jobs[j].next)
			})
			next := c.jobs[0]
			delay := next.next.Sub(now)
			if delay < 0 {
				delay = 0
			}
			timer = time.NewTimer(delay)
		}
		c.mutex.RUnlock()

		select {
		case <-timer.C:
			c.runJobs(time.Now().In(c.location))
		case <-c.stop:
			timer.Stop()
			return
		case newJob := <-c.add:
			timer.Stop()
			c.insertJob(newJob)
		case id := <-c.remove:
			timer.Stop()
			c.removeJob(id)
		}
	}
}

// runJobs executes all jobs that are scheduled for the given time
func (c *Cron) runJobs(now time.Time) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, job := range c.jobs {
		if job.next.After(now) || job.next.IsZero() {
			break
		}
		go job.job()
		job.next = job.schedule.Next(now)
	}
}

// insertJob inserts a new job into the job list
func (c *Cron) insertJob(job *jobEntry) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now().In(c.location)
	job.next = job.schedule.Next(now)
	c.jobs = append(c.jobs, job)
}

// Stop halts the cron scheduler
func (c *Cron) Stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.running {
		c.stop <- struct{}{}
		c.running = false
	}
}

// removeJob removes a job from the cron by its ID
func (c *Cron) removeJob(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var jobs []*jobEntry
	for _, job := range c.jobs {
		if job.id != id {
			jobs = append(jobs, job)
		}
	}
	c.jobs = jobs
}

// AddEverySecond adds a job to be executed every second
func (c *Cron) AddEverySecond(cmd func()) (string, error) {
	return c.AddFunc("* * * * * *", cmd)
}

// AddEveryNSeconds adds a job to be executed every N seconds
func (c *Cron) AddEveryNSeconds(n int, cmd func()) (string, error) {
	return c.AddFunc(fmt.Sprintf("*/%d * * * * *", n), cmd)
}

// AddEveryMinute adds a job to be executed every minute
func (c *Cron) AddEveryMinute(cmd func()) (string, error) {
	return c.AddFunc("0 * * * * *", cmd)
}

// AddEveryHour adds a job to be executed every hour
func (c *Cron) AddEveryHour(cmd func()) (string, error) {
	return c.AddFunc("0 0 * * * *", cmd)
}

// AddEveryDay adds a job to be executed every day at midnight
func (c *Cron) AddEveryDay(cmd func()) (string, error) {
	return c.AddFunc("0 0 0 * * *", cmd)
}

// AddEveryWeek adds a job to be executed every week on Sunday at midnight
func (c *Cron) AddEveryWeek(cmd func()) (string, error) {
	return c.AddFunc("0 0 0 * * 0", cmd)
}

// AddEveryMonth adds a job to be executed every month on the first day at midnight
func (c *Cron) AddEveryMonth(cmd func()) (string, error) {
	return c.AddFunc("0 0 0 1 * *", cmd)
}

// parseSchedule parses a cron schedule specification
func parseSchedule(spec string) (Schedule, error) {
	fields := strings.Fields(spec)
	if len(fields) != 6 {
		return nil, fmt.Errorf("invalid cron spec, expected 6 fields, got %d", len(fields))
	}

	schedule := &cronSchedule{
		second:     parseField(fields[0], 0, 59),
		minute:     parseField(fields[1], 0, 59),
		hour:       parseField(fields[2], 0, 23),
		dayOfMonth: parseField(fields[3], 1, 31),
		month:      parseField(fields[4], 1, 12),
		dayOfWeek:  parseField(fields[5], 0, 6),
	}

	return schedule, nil
}

// cronSchedule implements the Schedule interface
type cronSchedule struct {
	second, minute, hour, dayOfMonth, month, dayOfWeek []int
}

// Next returns the next activation time, later than the given time
func (s *cronSchedule) Next(t time.Time) time.Time {
	t = t.Add(time.Second)
	year, month, day := t.Date()
	hour, minute, second := t.Clock()

	// Find the next matching second
	second = s.nextValue(second, s.second)
	if second == s.second[0] {
		minute++
	}

	// Find the next matching minute
	minute = s.nextValue(minute, s.minute)
	if minute == s.minute[0] {
		hour++
	}

	// Find the next matching hour
	hour = s.nextValue(hour, s.hour)
	if hour == s.hour[0] {
		day++
	}

	for {
		// Check if the current day of the month and day of the week match
		if s.dayMatches(year, month, day) {
			return time.Date(year, month, day, hour, minute, second, 0, t.Location())
		}

		day++
		if day > daysIn(month, year) {
			day = 1
			month++
			if month > 12 {
				month = 1
				year++
			}
		}

		hour = s.hour[0]
		minute = s.minute[0]
		second = s.second[0]
	}
}

// dayMatches checks if the given date matches both the day-of-month and day-of-week constraints
func (s *cronSchedule) dayMatches(year int, month time.Month, day int) bool {
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return (len(s.dayOfMonth) == 0 || intSliceContains(s.dayOfMonth, day)) &&
		(len(s.dayOfWeek) == 0 || intSliceContains(s.dayOfWeek, int(t.Weekday())))
}

// nextValue returns the next value in the given slice that's larger than the given value
func (s *cronSchedule) nextValue(current int, values []int) int {
	for _, v := range values {
		if v > current {
			return v
		}
	}
	return values[0]
}

// parseField parses a cron field into a slice of integers
func parseField(field string, min, max int) []int {
	var result []int
	ranges := strings.Split(field, ",")
	for _, r := range ranges {
		if r == "*" {
			for i := min; i <= max; i++ {
				result = append(result, i)
			}
		} else if strings.Contains(r, "/") {
			parts := strings.Split(r, "/")
			step, _ := strconv.Atoi(parts[1])
			start := min
			if parts[0] != "*" {
				start, _ = strconv.Atoi(parts[0])
			}
			for i := start; i <= max; i += step {
				result = append(result, i)
			}
		} else if strings.Contains(r, "-") {
			parts := strings.Split(r, "-")
			start, _ := strconv.Atoi(parts[0])
			end, _ := strconv.Atoi(parts[1])
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
		} else {
			val, _ := strconv.Atoi(r)
			result = append(result, val)
		}
	}
	sort.Ints(result)
	return result
}

// daysIn returns the number of days in a month
func daysIn(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// intSliceContains checks if a slice contains a specific integer
func intSliceContains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
