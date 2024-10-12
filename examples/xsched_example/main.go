package main

import (
	"strings"
	"time"

	"github.com/seefs001/xox/xcolor"
	"github.com/seefs001/xox/xsched"
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

	// Wait for 35 seconds to allow jobs to run
	xcolor.Println(xcolor.White, "Waiting for 35 seconds...")
	start := time.Now()
	for time.Since(start) < 35*time.Second {
		time.Sleep(time.Second)
		elapsed := int(time.Since(start).Seconds())
		progressBar := xcolor.SprintMulti([]xcolor.ColorCode{xcolor.Purple}, "["+strings.Repeat("=", elapsed)+strings.Repeat(" ", 35-elapsed)+"]")
		xcolor.Print(xcolor.White, "\rElapsed: %2d seconds %s", elapsed, progressBar)
	}
	xcolor.Println(xcolor.White, "\n")

	// Remove the every-second job
	c.Remove(id1)
	xcolor.Println(xcolor.Yellow, "Job ID1 removed")

	// Wait for 5 more seconds
	time.Sleep(5 * time.Second)

	c.Stop()
	xcolor.Println(xcolor.Red, "Scheduler stopped")
}
