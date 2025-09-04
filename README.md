# dispatcher

`dispatcher` is an easy-to-use Go library designed to facilitate concurrent task processing through a managed worker pool. It allows efficient distribution of tasks across multiple worker goroutines, making it ideal for applications that require parallel execution of independent tasks such as processing jobs, or batching data processing tasks.

## Features

- **Scalable Worker Pool**: Manage the number of goroutines in the worker pool.
- **Graceful Shutdown**: Supports graceful shutdown to ensure that all in-progress jobs complete before the system stops.
- **Custom Job Processing**: Users can define custom jobs by implementing the Job interface.
- **Concurrency Control**: Control over concurrency settings to optimize performance according to hardware capabilities.

## Getting Started

### Prerequisites

- Go 1.22 or higher (https://golang.org/doc/install)

### Installing

To start using `dispatcher`, install the package using `go get`:

```bash
go get github.com/Brennan7/dispatcher
```

## Dispatcher Methods
- **New(maxWorkers int, queueSize int)**: Initializes a new Dispatcher with a specified number of workers and a custom job queue size. This allows for configuring the concurrency level (`maxWorkers`) and the capacity of pending jobs (`queueSize`) that can be buffered before blocking further submissions.
- **Start()**: Starts all workers and begins job processing.
- **AddJob(job Job)**: Adds a job to the dispatcher's job queue.
- **Stop()**: Stops all workers after they finish their current job and prevents new jobs from being added.

## Example Usage

For a practical example on how to set up and use the `dispatcher`, see the `main.go` file located in the `cmd/example` directory of this repository. This example demonstrates how to initialize the dispatcher, add jobs, and handle graceful shutdowns.

## License
dispatcher is released under the MIT License. See the LICENSE file for more information.