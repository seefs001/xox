package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/seefs001/xox/xcolor"
	"github.com/seefs001/xox/xsched"
)

// CronExample represents a cron job example
type CronExample struct {
	name        string
	spec        string
	description string
	color       xcolor.ColorCode
}

func main() {
	xcolor.Println(xcolor.Bold, "Starting Advanced Scheduler Example...")

	scheduler := xsched.New()

	// Define cron examples with different expressions
	examples := []CronExample{
		{
			name:        "Immediate and Every Second",
			spec:        "* * * * * *",
			description: "Runs immediately and then every second",
			color:       xcolor.Green,
		},
		{
			name:        "Immediate and Every 5 Seconds",
			spec:        "*/5 * * * * *",
			description: "Runs immediately and then every 5 seconds",
			color:       xcolor.Cyan,
		},
		{
			name:        "Every Minute",
			spec:        "0 * * * * *",
			description: "Runs at the start of every minute",
			color:       xcolor.Yellow,
		},
		{
			name:        "At 30 Seconds",
			spec:        "30 * * * * *",
			description: "Runs at 30th second of every minute",
			color:       xcolor.Blue,
		},
		{
			name:        "Work Hours",
			spec:        "0 0 9-17 * * 1-5",
			description: "Runs at the start of every hour during work hours (9-17) on weekdays",
			color:       xcolor.Purple,
		},
		{
			name:        "Multiple Times",
			spec:        "0,15,30,45 * * * * *",
			description: "Runs at 0,15,30,45 seconds of every minute",
			color:       xcolor.Red,
		},
	}

	// Print cron expression guide
	printCronGuide()

	// Add jobs and print their details
	xcolor.Println(xcolor.Bold, "\nConfigured Jobs:")
	xcolor.Println(xcolor.White, "%-25s %-20s %s", "Name", "Expression", "Description")
	xcolor.Println(xcolor.White, strings.Repeat("-", 80))

	jobIDs := make(map[string]string)
	for _, ex := range examples {
		id, err := scheduler.AddFuncWithOptions(ex.spec, createColoredJob(ex.name, ex.color), true)
		if err != nil {
			xcolor.Println(xcolor.Red, "Failed to add job '%s': %v", ex.name, err)
			continue
		}
		jobIDs[ex.name] = id
		xcolor.Println(ex.color, "%-25s %-20s %s", ex.name, ex.spec, ex.description)
	}

	// Start the scheduler
	scheduler.Start()
	xcolor.Println(xcolor.Bold, "\nScheduler started")

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start progress monitoring
	startTime := time.Now()
	go monitorProgress(startTime)

	xcolor.Println(xcolor.White, "\nPress Ctrl+C to exit")
	<-sigChan

	// Graceful shutdown
	xcolor.Println(xcolor.Yellow, "\n\nGracefully shutting down...")
	scheduler.Stop()
	xcolor.Println(xcolor.Red, "Scheduler stopped")
}

func createColoredJob(name string, color xcolor.ColorCode) func() {
	return func() {
		now := time.Now()
		xcolor.Println(color, "[%s] Executed at: %s.%03d",
			name,
			now.Format("15:04:05"),
			now.Nanosecond()/1000000)
	}
}

func monitorProgress(startTime time.Time) {
	for {
		time.Sleep(time.Second)
		elapsed := int(time.Since(startTime).Seconds())
		progressBar := createColoredProgressBar(elapsed, 60)
		xcolor.Print(xcolor.White, "\rRunning time: %2d seconds %s",
			elapsed,
			progressBar)
	}
}

func createColoredProgressBar(current, total int) string {
	progress := current % total
	bar := strings.Builder{}
	bar.WriteString("[")
	for i := 0; i < total; i++ {
		if i < progress {
			bar.WriteString("=")
		} else if i == progress {
			bar.WriteString(">")
		} else {
			bar.WriteString(" ")
		}
	}
	bar.WriteString("]")
	return xcolor.SprintMulti([]xcolor.ColorCode{xcolor.Purple}, bar.String())
}
func printCronGuide() {
	guide := `
Cron Expression Format:
┌──────── Second (0-59)
│ ┌────── Minute (0-59)
│ │ ┌──── Hour (0-23)
│ │ │ ┌── Day of Month (1-31)
│ │ │ │ ┌ Month (1-12)
│ │ │ │ │ ┌─ Day of Week (0-6) (Sunday=0)
* * * * * *

Special Characters:
* - Every unit
*/n - Every n units
n-m - Range from n to m
n,m,k - Specific values
n-m/k - Every k units between n-m
`
	xcolor.Println(xcolor.Cyan, guide)
}
