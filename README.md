# dispatcher

`dispatcher` is a Go library designed to facilitate concurrent task processing through a managed worker pool. It allows efficient distribution of tasks across multiple worker goroutines, making it ideal for applications that require parallel execution of independent tasks.

## Features

- Worker pool with configurable concurrency (maxWorkers)
- Bounded queue with backpressure (queueSize)
- Optional rate limiting:
    - FixedRateLimiter (evenly paced)
    - BurstRateLimiter (token bucket with bursts)
- Worker lifecycle hooks (OnJobStart, OnJobFinish)
- Graceful shutdown: Stop will wait for all in-flight jobs

## Getting Started

### Prerequisites

- Go 1.22 or higher (https://golang.org/doc/install)

### Installing

To start using `dispatcher`, install the package using `go get`:

```bash
go get github.com/Brennan7/dispatcher
```

## In a nutshell
Define a type that implements the Job interface.

Core API
- New(maxWorkers, queueSize, rateLimiter, hooks) → *Dispatcher
- Start() — spin up workers
- AddJob(job) — queue a job
- Stop() — wait for in-flight work and shut down

## Example Usage
See the examples under cmd/examples for a full setup.

## License
dispatcher is released under the MIT License. See the LICENSE file for more information.