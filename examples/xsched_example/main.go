package main

import (
	"fmt"
	"time"

	"github.com/seefs001/xox/xsched"
)

func Example() {
	// Create a new cron scheduler
	c := xsched.New()

	// Add a job that runs every minute
	id1, err := c.AddEveryMinute(func() {
		fmt.Println("This job runs every minute")
	})
	if err != nil {
		fmt.Printf("Error adding job: %v\n", err)
		return
	}

	// Add a job that runs every hour
	id2, err := c.AddEveryHour(func() {
		fmt.Println("This job runs every hour")
	})
	if err != nil {
		fmt.Printf("Error adding job: %v\n", err)
		return
	}

	// Add a job that runs every day at midnight
	id3, err := c.AddEveryDay(func() {
		fmt.Println("This job runs every day at midnight")
	})
	if err != nil {
		fmt.Printf("Error adding job: %v\n", err)
		return
	}

	// Start the scheduler
	c.Start()

	// Wait for some time to allow jobs to run
	time.Sleep(time.Minute * 2)

	// Remove a job
	c.Remove(id1)

	// Wait a bit more
	time.Sleep(time.Minute * 1)

	// Stop the scheduler
	c.Stop()

	fmt.Printf("Job IDs: %s, %s, %s\n", id1, id2, id3)

	// Output:
	// This job runs every minute
	// This job runs every minute
}

func ExampleCron_AddFunc() {
	c := xsched.New()

	id, err := c.AddFunc("*/5 * * * * *", func() {
		fmt.Println("This job runs every 5 seconds")
	})

	if err != nil {
		fmt.Printf("Error adding job: %v\n", err)
		return
	}

	c.Start()
	time.Sleep(time.Second * 11)
	c.Stop()

	fmt.Printf("Job ID: %s\n", id)

	// Output:
	// This job runs every 5 seconds
	// This job runs every 5 seconds
	// This job runs every 5 seconds
}

func main() {
	c := xsched.New()

	// Add a job that runs every second
	id, err := c.AddEverySecond(func() {
		fmt.Printf("Job executed at: %s\n", time.Now().Format("15:04:05.000"))
	})

	if err != nil {
		fmt.Printf("Error adding job: %v\n", err)
		return
	}

	fmt.Printf("Job added successfully with ID: %s\n", id)

	c.Start()
	fmt.Println("Scheduler started")

	// Wait for enough time to ensure the job executes multiple times
	fmt.Println("Waiting for 10 seconds...")
	start := time.Now()
	for time.Since(start) < 10*time.Second {
		time.Sleep(time.Second)
		fmt.Printf("Elapsed: %.2f seconds\n", time.Since(start).Seconds())
	}

	c.Remove(id)
	fmt.Println("Job removed")

	c.Stop()
	fmt.Println("Scheduler stopped")

	// Add a small delay to allow any pending output to be printed
	time.Sleep(time.Millisecond * 100)
}
