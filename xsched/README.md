# xsched

xsched is a flexible and efficient cron-like job scheduler for Go applications. It allows you to schedule and manage recurring tasks with ease, supporting both cron-style schedules and convenient preset intervals.

## Features

- Cron-style scheduling (supports seconds precision)
- Convenient methods for common scheduling patterns
- Custom tick interval for fine-grained control
- Concurrent job execution
- Dynamic job addition and removal
- Thread-safe operations

## Installation

To install xsched, use `go get`:

```bash
go get github.com/seefs001/xox/xsched
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/seefs001/xox/xsched"
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
5. Consider using a custom tick interval for high-precision scheduling.

## Performance Considerations

- The scheduler uses a single goroutine to manage timing, with each job running in its own goroutine.
- For high-precision scheduling (sub-second), use a smaller tick interval.
- Be mindful of resource usage when scheduling many frequent jobs.

## Thread Safety

The xsched package is designed to be thread-safe. You can add, remove, and manage jobs from multiple goroutines safely.

## Advanced Usage

### Custom Tick Interval

```go
c := xsched.NewWithTickInterval(50 * time.Millisecond)
```

### Handling Job Execution Errors

```go
c.AddFunc("* * * * * *", func() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered from panic:", r)
        }
    }()
    // Your job logic here
})
```

### Concurrent Job Execution

xsched automatically runs each job in its own goroutine, allowing for concurrent execution of multiple jobs.

## Examples

Here's an example demonstrating various features of xsched:

```go
package main

import (
    "fmt"
    "time"
    "github.com/seefs001/xox/xsched"
    "github.com/seefs001/xox/xcolor"
)

func main() {
    c := xsched.New()

    xcolor.Println(xcolor.Bold, "Starting Scheduler Example...")

    // Add a job that runs every second
    id1, _ := c.AddEverySecond(func() {
        xcolor.Println(xcolor.Green, "[Every Second] Executed at: %s", time.Now().Format("15:04:05"))
    })

    // Add a job that runs every 5 seconds
    id2, _ := c.AddFunc("*/5 * * * * *", func() {
        xcolor.Println(xcolor.Cyan, "[Every 5 Seconds] Executed at: %s", time.Now().Format("15:04:05"))
    })

    // Add a job that runs every minute
    id3, _ := c.AddEveryMinute(func() {
        xcolor.Println(xcolor.Yellow, "[Every Minute] Executed at: %s", time.Now().Format("15:04:05"))
    })

    // Add a job that runs at 30 seconds past each minute
    id4, _ := c.AddFunc("30 * * * * *", func() {
        xcolor.Println(xcolor.Blue, "[At 30 Seconds] Executed at: %s", time.Now().Format("15:04:05"))
    })

    xcolor.Println(xcolor.White, "Jobs added successfully with IDs:")
    xcolor.Println(xcolor.White, "ID1: %s", id1)
    xcolor.Println(xcolor.White, "ID2: %s", id2)
    xcolor.Println(xcolor.White, "ID3: %s", id3)
    xcolor.Println(xcolor.White, "ID4: %s", id4)

    c.Start()
    xcolor.Println(xcolor.Bold, "Scheduler started")

    // Run for 35 seconds
    time.Sleep(35 * time.Second)

    // Remove the every-second job
    c.Remove(id1)
    xcolor.Println(xcolor.Yellow, "Job ID1 removed")

    // Run for 5 more seconds
    time.Sleep(5 * time.Second)

    c.Stop()
    xcolor.Println(xcolor.Red, "Scheduler stopped")
}
```

This example demonstrates adding jobs with different schedules, using convenience methods, removing a job, and stopping the scheduler.

## Contributing

Contributions to xsched are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
