package main

import (
	"fmt"
	"time"

	"github.com/Brennan7/dispatcher"
)

// Example: Minimal usage (no rate limiting, no hooks)
// Processes jobs as fast as possible for maximum throughput

type PrintJob struct {
	Message string
}

// Process Implements the Job interface.
func (p PrintJob) Process() error {
	fmt.Println("Processing:", p.Message)
	return nil
}

func main() {
	// Create the dispatcher:
	// - maxWorkers: number of worker goroutines
	// - queueSize: buffered job capacity (AddJob blocks when full)
	// - rateLimiter: nil = no rate limiting (fastest possible)
	// - hooks: nil = no lifecycle callbacks
	d := dispatcher.New(
		4,
		32,
		nil,
		nil,
	)

	d.Start()

	// Enqueue some jobs; they will run as fast as workers can process them
	for i := 1; i <= 20; i++ {
		d.AddJob(PrintJob{Message: fmt.Sprintf("job %d", i)})
	}

	// Let jobs process for a short while
	time.Sleep(1 * time.Second)

	// Graceful shutdown: lets in-flight jobs finish
	d.Stop()
}
