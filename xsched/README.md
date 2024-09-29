# xsched

xsched is a flexible and efficient cron-like job scheduler for Go applications. It allows you to schedule and manage recurring tasks with ease, supporting both cron-style schedules and convenient preset intervals.

## Features

- Cron-style scheduling (supports seconds precision)
- Convenient methods for common scheduling patterns
- Custom tick interval for fine-grained control
- Concurrent job execution
- Dynamic job addition and removal

## Installation

To install xsched, use `go get`:

```bash
go get github.com/yourusername/xsched
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/yourusername/xsched"
    "time"
)

func main() {
    // Create a new cron scheduler
    c := xsched.New()

    // Add a job that runs every second
    c.AddFunc("* * * * * *", func() {
        fmt.Println("This job runs every second")
    })

    // Start the scheduler
    c.Start()

    // Run for 5 seconds
    time.Sleep(5 * time.Second)

    // Stop the scheduler
    c.Stop()
}
```

### API Reference

#### Creating a Scheduler

```go
// Create a new scheduler with default tick interval (1 second)
c := xsched.New()

// Create a new scheduler with custom tick interval
c := xsched.NewWithTickInterval(100 * time.Millisecond)
```

#### Adding Jobs

```go
// Add a job with a cron schedule
id, err := c.AddFunc("0 */5 * * * *", func() {
    fmt.Println("This job runs every 5 minutes")
})

// Convenience methods for common intervals
c.AddEverySecond(func() { fmt.Println("Every second") })
c.AddEveryNSeconds(30, func() { fmt.Println("Every 30 seconds") })
c.AddEveryMinute(func() { fmt.Println("Every minute") })
c.AddEveryHour(func() { fmt.Println("Every hour") })
c.AddEveryDay(func() { fmt.Println("Every day at midnight") })
c.AddEveryWeek(func() { fmt.Println("Every week on Sunday at midnight") })
c.AddEveryMonth(func() { fmt.Println("Every month on the first day at midnight") })
```

#### Managing the Scheduler

```go
// Start the scheduler
c.Start()

// Stop the scheduler
c.Stop()

// Remove a job by its ID
c.Remove(id)
```

### Cron Schedule Format

The cron schedule format is as follows:

```
┌─────────────── second (0 - 59)
│ ┌───────────── minute (0 - 59)
│ │ ┌─────────── hour (0 - 23)
│ │ │ ┌───────── day of month (1 - 31)
│ │ │ │ ┌─────── month (1 - 12)
│ │ │ │ │ ┌───── day of week (0 - 6) (Sunday to Saturday)
│ │ │ │ │ │
* * * * * *
```

Special characters:
- `*`: any value
- `,`: value list separator
- `-`: range of values
- `/`: step values

Example: `*/15 * * * * *` runs every 15 seconds.

## Best Practices

1. Use appropriate tick intervals: Smaller intervals provide more precision but consume more resources.
2. Avoid long-running jobs: Keep jobs short to prevent blocking other scheduled tasks.
3. Handle errors in your job functions to prevent crashes.
4. Use `Remove()` to clean up jobs that are no longer needed.

## Performance Considerations

- The scheduler uses a single goroutine to manage timing, with each job running in its own goroutine.
- For high-precision scheduling (sub-second), use a smaller tick interval.
- Be mindful of resource usage when scheduling many frequent jobs.

## Thread Safety

The xsched package is designed to be thread-safe. You can add, remove, and manage jobs from multiple goroutines safely.
