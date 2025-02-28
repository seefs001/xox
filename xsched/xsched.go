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
	mutex        sync.RWMutex
	location     *time.Location
	tickInterval time.Duration
	errorHandler func(error)
}

// jobEntry represents a scheduled job
type jobEntry struct {
	schedule Schedule
	job      Job
	next     time.Time
	id       string
	runOnAdd bool
}

// CronField represents a single field in a cron expression
type CronField struct {
	Min   int
	Max   int
	Valid []int
}

// CronExpression represents a parsed cron expression
type CronExpression struct {
	Second     CronField
	Minute     CronField
	Hour       CronField
	DayOfMonth CronField
	Month      CronField
	DayOfWeek  CronField
}

// CronBuilder helps build cron expressions using method chaining
type CronBuilder struct {
	expr *CronExpression
	err  error
}

// New creates a new Cron instance with default tick interval of 1 second
func New() *Cron {
	return &Cron{
		jobs:         make([]*jobEntry, 0, 10),
		location:     time.Local,
		tickInterval: time.Second,
		errorHandler: func(err error) {},
	}
}

// NewWithTickInterval creates a new Cron instance with a custom tick interval
func NewWithTickInterval(interval time.Duration) *Cron {
	if interval < time.Millisecond {
		interval = time.Second
	}
	return &Cron{
		jobs:         make([]*jobEntry, 0, 10),
		location:     time.Local,
		tickInterval: interval,
		errorHandler: func(err error) {},
	}
}

// AddFunc adds a function to be executed on the given schedule
func (c *Cron) AddFunc(spec string, cmd func()) (string, error) {
	return c.AddFuncWithOptions(spec, cmd, false)
}

// AddFuncWithOptions adds a function with additional options
func (c *Cron) AddFuncWithOptions(spec string, cmd func(), runOnAdd bool) (string, error) {
	schedule, err := parseSchedule(spec)
	if err != nil {
		return "", err
	}
	return c.addJobWithOptions(schedule, Job(cmd), runOnAdd), nil
}

// addJobWithOptions adds a job with additional options
func (c *Cron) addJobWithOptions(schedule Schedule, cmd Job, runOnAdd bool) string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	job := &jobEntry{
		schedule: schedule,
		job:      cmd,
		id:       fmt.Sprintf("%p", cmd),
		runOnAdd: runOnAdd,
	}

	now := time.Now().In(c.location)
	if runOnAdd {
		go job.job()
	}
	job.next = schedule.Next(now)
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

// StartBlocking begins the cron scheduler and blocks until Stop is called
func (c *Cron) StartBlocking() {
	c.mutex.Lock()
	if c.running {
		c.mutex.Unlock()
		return
	}
	c.running = true
	c.mutex.Unlock()

	c.run()
}

// Start begins the cron scheduler in non-blocking mode
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
	ticker := time.NewTicker(c.tickInterval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now().In(c.location)
		var jobsToRun []*jobEntry
		var jobsToUpdate map[*jobEntry]time.Time

		// First, get jobs that need to run with a read lock
		c.mutex.RLock()
		if !c.running {
			c.mutex.RUnlock()
			return
		}

		jobsToUpdate = make(map[*jobEntry]time.Time)

		for _, job := range c.jobs {
			if job.next.IsZero() {
				nextRun := job.schedule.Next(now)
				jobsToUpdate[job] = nextRun
				continue
			}

			if now.After(job.next) || now.Equal(job.next) {
				// Create a copy to avoid data races
				jobCopy := *job // Make a copy of the job struct
				jobsToRun = append(jobsToRun, &jobCopy)

				// Track the job that needs its next execution time updated
				nextRun := job.schedule.Next(now)
				jobsToUpdate[job] = nextRun
			}
		}
		c.mutex.RUnlock()

		// Update the next execution times with a write lock
		if len(jobsToUpdate) > 0 {
			c.mutex.Lock()
			for job, nextTime := range jobsToUpdate {
				job.next = nextTime
			}
			c.mutex.Unlock()
		}

		// Execute jobs without holding any lock
		for _, job := range jobsToRun {
			go func(j *jobEntry) {
				defer func() {
					if r := recover(); r != nil {
						if c.errorHandler != nil {
							c.errorHandler(fmt.Errorf("job panic: %v", r))
						}
					}
				}()
				j.job()
			}(job)
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
	// Start with the next second
	t = t.Add(time.Second)

	// Reset sub-second nanoseconds
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())

	// Find the next matching time
	for {
		if !s.matchField(s.month, int(t.Month())) {
			t = time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
			continue
		}
		if !s.matchField(s.dayOfMonth, t.Day()) || !s.matchField(s.dayOfWeek, int(t.Weekday())) {
			t = t.AddDate(0, 0, 1)
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
			continue
		}
		if !s.matchField(s.hour, t.Hour()) {
			t = t.Add(time.Hour)
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
			continue
		}
		if !s.matchField(s.minute, t.Minute()) {
			t = t.Add(time.Minute)
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
			continue
		}
		if !s.matchField(s.second, t.Second()) {
			t = t.Add(time.Second)
			continue
		}
		break
	}
	return t
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

// AddEveryNMinutes adds a job to be executed every N minutes
func (c *Cron) AddEveryNMinutes(n int, cmd func()) (string, error) {
	if n <= 0 || n > 59 {
		return "", fmt.Errorf("invalid interval: %d", n)
	}
	spec := fmt.Sprintf("0 */%d * * * *", n)
	return c.AddFunc(spec, cmd)
}

// AddEveryNHours adds a job to be executed every N hours
func (c *Cron) AddEveryNHours(n int, cmd func()) (string, error) {
	if n <= 0 || n > 23 {
		return "", fmt.Errorf("invalid interval: %d", n)
	}
	spec := fmt.Sprintf("0 0 */%d * * *", n)
	return c.AddFunc(spec, cmd)
}

// GetJobCount returns the number of jobs in the cron scheduler
func (c *Cron) GetJobCount() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.jobs)
}

// Clear removes all jobs from the cron scheduler
func (c *Cron) Clear() {
	c.mutex.Lock()
	c.jobs = make([]*jobEntry, 0, 10)
	c.mutex.Unlock()
}

// SetErrorHandler sets the error handler for the cron scheduler
func (c *Cron) SetErrorHandler(eh func(error)) {
	c.mutex.Lock()
	c.errorHandler = eh
	c.mutex.Unlock()
}

// SetLocation sets the location for the cron scheduler
func (c *Cron) SetLocation(loc *time.Location) {
	if loc == nil {
		loc = time.Local
	}
	c.mutex.Lock()
	c.location = loc
	c.mutex.Unlock()
}

// NewCronExpression creates a new CronExpression with default field ranges
func NewCronExpression() *CronExpression {
	return &CronExpression{
		Second:     CronField{Min: 0, Max: 59},
		Minute:     CronField{Min: 0, Max: 59},
		Hour:       CronField{Min: 0, Max: 23},
		DayOfMonth: CronField{Min: 1, Max: 31},
		Month:      CronField{Min: 1, Max: 12},
		DayOfWeek:  CronField{Min: 0, Max: 6},
	}
}

// ParseExpression parses a cron expression string into a CronExpression
func ParseExpression(spec string) (*CronExpression, error) {
	fields := strings.Fields(spec)
	if len(fields) != 6 {
		return nil, fmt.Errorf("invalid cron expression: expected 6 fields, got %d", len(fields))
	}

	expr := NewCronExpression()
	fieldParsers := []struct {
		field *CronField
		value string
	}{
		{&expr.Second, fields[0]},
		{&expr.Minute, fields[1]},
		{&expr.Hour, fields[2]},
		{&expr.DayOfMonth, fields[3]},
		{&expr.Month, fields[4]},
		{&expr.DayOfWeek, fields[5]},
	}

	for _, fp := range fieldParsers {
		valid, err := parseField(fp.value, fp.field.Min, fp.field.Max)
		if err != nil {
			return nil, err
		}
		fp.field.Valid = valid
	}

	return expr, nil
}

// ValidateExpression validates a cron expression string
func ValidateExpression(spec string) error {
	_, err := ParseExpression(spec)
	return err
}

// BuildSchedule creates a Schedule from a CronExpression
func BuildSchedule(expr *CronExpression) Schedule {
	return &cronSchedule{
		second:     expr.Second.Valid,
		minute:     expr.Minute.Valid,
		hour:       expr.Hour.Valid,
		dayOfMonth: expr.DayOfMonth.Valid,
		month:      expr.Month.Valid,
		dayOfWeek:  expr.DayOfWeek.Valid,
	}
}

// BuildScheduleFromSpec creates a Schedule directly from a cron expression string
func BuildScheduleFromSpec(spec string) (Schedule, error) {
	expr, err := ParseExpression(spec)
	if err != nil {
		return nil, err
	}
	return BuildSchedule(expr), nil
}

// IsValid checks if a specific time matches the cron expression
func (expr *CronExpression) IsValid(t time.Time) bool {
	return contains(expr.Second.Valid, t.Second()) &&
		contains(expr.Minute.Valid, t.Minute()) &&
		contains(expr.Hour.Valid, t.Hour()) &&
		contains(expr.DayOfMonth.Valid, t.Day()) &&
		contains(expr.Month.Valid, int(t.Month())) &&
		contains(expr.DayOfWeek.Valid, int(t.Weekday()))
}

// contains checks if a slice contains a specific value
func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// NewCronBuilder creates a new CronBuilder instance
func NewCronBuilder() *CronBuilder {
	return &CronBuilder{
		expr: NewCronExpression(),
	}
}

// WithSeconds sets the seconds field
func (b *CronBuilder) WithSeconds(seconds ...int) *CronBuilder {
	if b.err != nil {
		return b
	}
	if err := validateRange(seconds, b.expr.Second.Min, b.expr.Second.Max); err != nil {
		b.err = fmt.Errorf("invalid seconds: %v", err)
		return b
	}
	b.expr.Second.Valid = seconds
	return b
}

// WithMinutes sets the minutes field
func (b *CronBuilder) WithMinutes(minutes ...int) *CronBuilder {
	if b.err != nil {
		return b
	}
	if err := validateRange(minutes, b.expr.Minute.Min, b.expr.Minute.Max); err != nil {
		b.err = fmt.Errorf("invalid minutes: %v", err)
		return b
	}
	b.expr.Minute.Valid = minutes
	return b
}

// WithHours sets the hours field
func (b *CronBuilder) WithHours(hours ...int) *CronBuilder {
	if b.err != nil {
		return b
	}
	if err := validateRange(hours, b.expr.Hour.Min, b.expr.Hour.Max); err != nil {
		b.err = fmt.Errorf("invalid hours: %v", err)
		return b
	}
	b.expr.Hour.Valid = hours
	return b
}

// WithDaysOfMonth sets the days of month field
func (b *CronBuilder) WithDaysOfMonth(days ...int) *CronBuilder {
	if b.err != nil {
		return b
	}
	if err := validateRange(days, b.expr.DayOfMonth.Min, b.expr.DayOfMonth.Max); err != nil {
		b.err = fmt.Errorf("invalid days of month: %v", err)
		return b
	}
	b.expr.DayOfMonth.Valid = days
	return b
}

// WithMonths sets the months field
func (b *CronBuilder) WithMonths(months ...int) *CronBuilder {
	if b.err != nil {
		return b
	}
	if err := validateRange(months, b.expr.Month.Min, b.expr.Month.Max); err != nil {
		b.err = fmt.Errorf("invalid months: %v", err)
		return b
	}
	b.expr.Month.Valid = months
	return b
}

// WithDaysOfWeek sets the days of week field
func (b *CronBuilder) WithDaysOfWeek(days ...int) *CronBuilder {
	if b.err != nil {
		return b
	}
	if err := validateRange(days, b.expr.DayOfWeek.Min, b.expr.DayOfWeek.Max); err != nil {
		b.err = fmt.Errorf("invalid days of week: %v", err)
		return b
	}
	b.expr.DayOfWeek.Valid = days
	return b
}

// WithEverySecond sets the expression to run every second
func (b *CronBuilder) WithEverySecond() *CronBuilder {
	seconds := make([]int, b.expr.Second.Max-b.expr.Second.Min+1)
	for i := range seconds {
		seconds[i] = b.expr.Second.Min + i
	}
	return b.WithSeconds(seconds...)
}

// WithEveryMinute sets the expression to run every minute at second 0
func (b *CronBuilder) WithEveryMinute() *CronBuilder {
	minutes := make([]int, b.expr.Minute.Max-b.expr.Minute.Min+1)
	for i := range minutes {
		minutes[i] = b.expr.Minute.Min + i
	}
	return b.WithSeconds(0).WithMinutes(minutes...)
}

// WithEveryHour sets the expression to run every hour at minute 0, second 0
func (b *CronBuilder) WithEveryHour() *CronBuilder {
	hours := make([]int, b.expr.Hour.Max-b.expr.Hour.Min+1)
	for i := range hours {
		hours[i] = b.expr.Hour.Min + i
	}
	return b.WithSeconds(0).WithMinutes(0).WithHours(hours...)
}

// WithEveryDay sets the expression to run every day at 00:00:00
func (b *CronBuilder) WithEveryDay() *CronBuilder {
	return b.WithSeconds(0).WithMinutes(0).WithHours(0)
}

// WithInterval sets an interval for a specific field
func (b *CronBuilder) WithInterval(field string, start, interval int) *CronBuilder {
	if b.err != nil {
		return b
	}

	var values []int
	var min, max int

	switch strings.ToLower(field) {
	case "second":
		min, max = b.expr.Second.Min, b.expr.Second.Max
	case "minute":
		min, max = b.expr.Minute.Min, b.expr.Minute.Max
	case "hour":
		min, max = b.expr.Hour.Min, b.expr.Hour.Max
	default:
		b.err = fmt.Errorf("invalid field: %s", field)
		return b
	}

	if start < min || start > max {
		b.err = fmt.Errorf("invalid start value %d for field %s", start, field)
		return b
	}

	for i := start; i <= max; i += interval {
		values = append(values, i)
	}

	switch strings.ToLower(field) {
	case "second":
		return b.WithSeconds(values...)
	case "minute":
		return b.WithMinutes(values...)
	case "hour":
		return b.WithHours(values...)
	}

	return b
}

// Build creates a Schedule from the builder
func (b *CronBuilder) Build() (Schedule, error) {
	if b.err != nil {
		return nil, b.err
	}
	return BuildSchedule(b.expr), nil
}

// validateRange checks if all values are within the specified range
func validateRange(values []int, min, max int) error {
	if len(values) == 0 {
		return fmt.Errorf("no values provided")
	}
	for _, v := range values {
		if v < min || v > max {
			return fmt.Errorf("value %d out of range [%d,%d]", v, min, max)
		}
	}
	return nil
}

// String returns the cron expression as a string
func (b *CronBuilder) String() (string, error) {
	if b.err != nil {
		return "", b.err
	}

	expr := b.expr
	return fmt.Sprintf("%v %v %v %v %v %v",
		formatField(expr.Second.Valid),
		formatField(expr.Minute.Valid),
		formatField(expr.Hour.Valid),
		formatField(expr.DayOfMonth.Valid),
		formatField(expr.Month.Valid),
		formatField(expr.DayOfWeek.Valid)), nil
}

// formatField converts a slice of integers to a cron field string
func formatField(values []int) string {
	if len(values) == 0 {
		return "*"
	}
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(values)), ","), "[]")
}
