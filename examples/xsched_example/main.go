package main

import (
	"fmt"
	"time"

	"github.com/seefs001/xox/xsched"
)

func main() {
	c := xsched.New()

	// Add a job that runs every second
	id1, err := c.AddEverySecond(func() {
		fmt.Printf("Job executed every second at: %s\n", time.Now().Format("15:04:05.000"))
	})
	if err != nil {
		fmt.Printf("Error adding job: %v\n", err)
		return
	}

	// Add a job that runs every 5 seconds
	id2, err := c.AddFunc("*/5 * * * * *", func() {
		fmt.Println("This job runs every 5 seconds")
	})
	if err != nil {
		fmt.Printf("Error adding job: %v\n", err)
		return
	}

	fmt.Printf("Jobs added successfully with IDs: %s, %s\n", id1, id2)

	c.Start()
	fmt.Println("Scheduler started")

	// Wait for 15 seconds to allow jobs to run
	fmt.Println("Waiting for 15 seconds...")
	start := time.Now()
	for time.Since(start) < 15*time.Second {
		time.Sleep(time.Second)
		fmt.Printf("Elapsed: %.2f seconds\n", time.Since(start).Seconds())
	}

	// Remove the first job
	c.Remove(id1)
	fmt.Println("Job 1 removed")

	// Wait for 5 more seconds
	time.Sleep(5 * time.Second)

	c.Stop()
	fmt.Println("Scheduler stopped")

	// Add a small delay to allow any pending output to be printed
	time.Sleep(time.Millisecond * 100)
}
