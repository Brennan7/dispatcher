package dispatcher

import (
	"sync"
)

// Dispatcher manages the distribution of jobs to available workers
type Dispatcher struct {
	jobQueue    chan Job       // A buffered channel that holds submitted jobs
	workerPool  chan chan Job  // A pool of worker channels that are ready to receive jobs
	maxWorkers  int            // The maximum number of workers
	rateLimiter RateLimiter    // Optional rate limiter for controlling job processing rate. If nil, jobs are processed as fast as possible
	hooks       *WorkerHooks   // Optional hooks for worker logging/monitoring
	wg          sync.WaitGroup // WaitGroup to track the completion of all jobs
}

// New initializes a Dispatcher with workers, queue size, optional rate limiter and hooks
func New(maxWorkers, queueSize int, rateLimiter RateLimiter, hooks *WorkerHooks) *Dispatcher {
	jobQueue := make(chan Job, queueSize)         // Initialize the job queue with a user-defined buffer size
	workerPool := make(chan chan Job, maxWorkers) // Create a channel of job channels with a capacity equal to the max number of workers
	return &Dispatcher{
		jobQueue:    jobQueue,
		workerPool:  workerPool,
		maxWorkers:  maxWorkers,
		rateLimiter: rateLimiter,
		hooks:       hooks,
	}
}

// Start initializes the workers and starts the dispatching process
func (d *Dispatcher) Start() {
	for i := 1; i <= d.maxWorkers; i++ {
		worker := newWorker(i, d.workerPool, d.hooks) // Initialize a new worker
		worker.start()                                // start the worker in a goroutine where it waits to receive jobs
	}

	go d.dispatch() // start the main dispatch loop in a separate goroutine
}

// AddJob adds a new job to the dispatcher's job queue
func (d *Dispatcher) AddJob(job Job) {
	d.wg.Add(1)       // Increment the WaitGroup counter before adding a job to handle it properly in dispatch
	d.jobQueue <- job // Send the job to the job queue
}

// Stop initiates a graceful shutdown
func (d *Dispatcher) Stop() {
	close(d.jobQueue) // Close the job queue to prevent any new jobs from being added
	d.wg.Wait()       // Block until all jobs have been processed

	// Close all worker job channels to signal them to stop after all jobs are processed
	for i := 0; i < d.maxWorkers; i++ {
		jobChannel := <-d.workerPool // Retrieve each worker's job channel
		close(jobChannel)            // Close the channel to signal the worker to terminate
	}
}

// dispatch continuously takes jobs from jobQueue and sends them to available workers
func (d *Dispatcher) dispatch() {
	for job := range d.jobQueue { // Continuously receive jobs from the job queue
		if d.rateLimiter != nil {
			d.rateLimiter.Acquire()
		}

		jobChannel := <-d.workerPool // Wait for an available worker's job channel
		jobChannel <- job            // Send the job to the retrieved worker's job channel
		d.wg.Done()                  // Signal that the job has been dispatched and is now being handled by a worker
	}
}
