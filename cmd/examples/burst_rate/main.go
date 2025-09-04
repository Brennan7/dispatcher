package main

import (
	"fmt"
	"time"

	"github.com/Brennan7/dispatcher"
)

// Example: Burst rate limiting
// Allows short bursts while keeping average throughput near the limit

type PrintJob struct {
	Message string
}

// Process Implements the Job interface.
func (p PrintJob) Process() error {
	fmt.Println("Processing (burst):", p.Message)
	return nil
}

func main() {
	// Create a burst rate limiter (e.g., ~25 jobs/second on average)
	limiter := dispatcher.NewBurstRateLimiter(25)

	// Create the dispatcher:
	// - maxWorkers: number of worker goroutines
	// - queueSize: buffered job capacity
	// - rateLimiter: burst limiter
	// - hooks: nil (no lifecycle callbacks for simplicity)
	d := dispatcher.New(
		8,
		256,
		limiter,
		nil,
	)

	d.Start()

	// Enqueue a burst of jobs
	for i := 1; i <= 200; i++ {
		d.AddJob(PrintJob{Message: fmt.Sprintf("job %d", i)})
	}

	// Let jobs process for a short while
	time.Sleep(5 * time.Second)

	// Graceful shutdown: lets in-flight jobs finish
	d.Stop()
}
