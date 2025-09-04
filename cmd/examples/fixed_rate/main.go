package main

import (
	"fmt"
	"time"

	"github.com/Brennan7/dispatcher"
)

// Example: Fixed rate limiting
// Evenly spaces job processing at a target rate (e.g., 100 jobs/second)

type PrintJob struct {
	Message string
}

// Process Implements the Job interface
func (p PrintJob) Process() error {
	fmt.Println("Processing (fixed):", p.Message)
	return nil
}

func main() {
	// Create a fixed rate limiter: dispatch jobs at a steady pace.
	fixed := dispatcher.NewFixedRateLimiter(100) // 100 jobs/second

	// Create the dispatcher:
	// - maxWorkers: number of worker goroutines
	// - queueSize: buffered job capacity
	// - rateLimiter: fixed limiter for evenly spaced dispatch
	// - hooks: nil (no lifecycle callbacks for simplicity)
	d := dispatcher.New(
		4,
		100,
		fixed,
		nil,
	)

	d.Start()

	// Enqueue jobs; they will be processed at a steady, fixed pace.
	for i := 1; i <= 100; i++ {
		d.AddJob(PrintJob{Message: fmt.Sprintf("job %d", i)})
	}

	// Let jobs process for a short while.
	time.Sleep(1 * time.Second)

	// Graceful shutdown: lets in-flight jobs finish.
	d.Stop()
}
