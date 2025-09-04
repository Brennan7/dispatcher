package dispatcher

// worker represents a concurrent worker that processes jobs
type worker struct {
	id         int           // Unique identifier for the worker
	workerPool chan chan Job // A channel on which this worker will send its job channel when ready
	jobChannel chan Job      // A channel for receiving jobs to process
	hooks      *WorkerHooks  // Optional hooks
}

// WorkerHooks allows users to hook into worker lifecycle events
type WorkerHooks struct {
	OnJobStart  func(workerID int, job Job)
	OnJobFinish func(workerID int, job Job, err error)
}

// newWorker initializes a new instance of a Worker
func newWorker(id int, workerPool chan chan Job, hooks *WorkerHooks) worker {
	return worker{
		id:         id,
		workerPool: workerPool,
		jobChannel: make(chan Job), // Initialize the job channel for receiving jobs
		hooks:      hooks,
	}
}

// start begins the worker's loop in a new goroutine, waiting for jobs and quit signals
func (w worker) start() {
	go func() {
		for {
			// The worker registers its jobChannel to the workerPool, signaling it's ready for work
			// This operation can block if the pool is currently full and no dispatcher is taking channels out
			w.workerPool <- w.jobChannel

			select {
			case job, ok := <-w.jobChannel:
				if !ok {
					return
				}

				if w.hooks != nil && w.hooks.OnJobStart != nil {
					w.hooks.OnJobStart(w.id, job)
				}

				err := job.Process()

				if w.hooks != nil && w.hooks.OnJobFinish != nil {
					w.hooks.OnJobFinish(w.id, job, err)
				}
			}
		}
	}()
}
