package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/Brennan7/dispatcher"
)

// Example: Using WorkerHooks (lifecycle logging)
// Demonstrates OnJobStart and OnJobFinish callbacks for observability

type PrintJob struct {
	Message string
	Fail    bool // set true to simulate a processing error
}

// Process Implements the Job interface.
func (p PrintJob) Process() error {
	if p.Fail {
		return errors.New("simulated failure")
	}
	fmt.Println("Processing:", p.Message)
	return nil
}

func main() {
	// Create hooks to observe job lifecycle events
	hooks := &dispatcher.WorkerHooks{
		OnJobStart: func(workerID int, job dispatcher.Job) {
			// This is called right before Process() is executed
			if pj, ok := job.(PrintJob); ok {
				fmt.Printf("[worker %d] start: %s (fail=%v)\n", workerID, pj.Message, pj.Fail)
			} else {
				fmt.Printf("[worker %d] start: %#v\n", workerID, job)
			}
		},
		OnJobFinish: func(workerID int, job dispatcher.Job, err error) {
			// This is called after Process() returns (with success or error)
			if err != nil {
				fmt.Printf("[worker %d] finish: ERROR: %v\n", workerID, err)
				return
			}
			if pj, ok := job.(PrintJob); ok {
				fmt.Printf("[worker %d] finish: %s OK\n", workerID, pj.Message)
			} else {
				fmt.Printf("[worker %d] finish: %#v OK\n", workerID, job)
			}
		},
	}

	// Create the dispatcher:
	// - maxWorkers: number of worker goroutines
	// - queueSize: buffered job capacity
	// - rateLimiter: nil = no rate limiting
	// - hooks: provide lifecycle callbacks for logging/metrics
	d := dispatcher.New(
		4,
		32,
		nil,
		hooks,
	)

	d.Start()

	// Enqueue jobs; every 4th job will fail to demonstrate error handling in hooks
	for i := 1; i <= 12; i++ {
		d.AddJob(PrintJob{
			Message: fmt.Sprintf("job %d", i),
			Fail:    i%4 == 0, // simulate periodic failures
		})
	}

	// Let jobs process for a short while.
	time.Sleep(2 * time.Second)

	// Graceful shutdown: lets in-flight jobs finish
	d.Stop()
}
