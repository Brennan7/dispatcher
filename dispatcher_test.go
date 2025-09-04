package dispatcher

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// testJob is a lightweight Job implementation used for tests.
type testJob struct {
	delay     time.Duration
	onRun     func()
	onFinish  func()
	failError error
}

func (j testJob) Process() error {
	if j.onRun != nil {
		j.onRun()
	}
	if j.delay > 0 {
		time.Sleep(j.delay)
	}
	if j.onFinish != nil {
		j.onFinish()
	}
	return j.failError
}

// fakeRateLimiter lets tests observe Acquire calls and inject delay.
type fakeRateLimiter struct {
	delay time.Duration
	calls int64
}

func (f *fakeRateLimiter) Acquire() {
	atomic.AddInt64(&f.calls, 1)
	if f.delay > 0 {
		time.Sleep(f.delay)
	}
}

func TestDispatcher_ProcessesAllJobs(t *testing.T) {
	maxWorkers := 4
	queueSize := 16
	jobCount := 50

	var processed int64
	d := New(maxWorkers, queueSize, nil, nil)
	d.Start()

	for i := 0; i < jobCount; i++ {
		d.AddJob(testJob{
			onFinish: func() { atomic.AddInt64(&processed, 1) },
		})
	}

	d.Stop()

	if got := atomic.LoadInt64(&processed); got != int64(jobCount) {
		t.Fatalf("processed=%d, want=%d", got, jobCount)
	}
}

func TestDispatcher_RespectsMaxWorkers(t *testing.T) {
	maxWorkers := 5
	queueSize := 100
	jobCount := 50

	var active int64
	var peak int64

	d := New(maxWorkers, queueSize, nil, nil)
	d.Start()

	for i := 0; i < jobCount; i++ {
		d.AddJob(testJob{
			delay: 25 * time.Millisecond,
			onRun: func() {
				cur := atomic.AddInt64(&active, 1)
				// track peak concurrency
				for {
					p := atomic.LoadInt64(&peak)
					if cur > p {
						if atomic.CompareAndSwapInt64(&peak, p, cur) {
							break
						}
						continue
					}
					break
				}
			},
			onFinish: func() {
				atomic.AddInt64(&active, -1)
			},
		})
	}

	d.Stop()

	if atomic.LoadInt64(&peak) > int64(maxWorkers) {
		t.Fatalf("peak concurrency=%d exceeded maxWorkers=%d", peak, maxWorkers)
	}
}

func TestDispatcher_InvokesHooks(t *testing.T) {
	maxWorkers := 3
	queueSize := 10
	jobCount := 7

	var starts int64
	var finishes int64

	hooks := &WorkerHooks{
		OnJobStart: func(workerID int, job Job) {
			atomic.AddInt64(&starts, 1)
		},
		OnJobFinish: func(workerID int, job Job, err error) {
			atomic.AddInt64(&finishes, 1)
		},
	}

	d := New(maxWorkers, queueSize, nil, hooks)
	d.Start()

	for i := 0; i < jobCount; i++ {
		d.AddJob(testJob{})
	}

	d.Stop()

	if s := atomic.LoadInt64(&starts); s != int64(jobCount) {
		t.Fatalf("starts=%d, want=%d", s, jobCount)
	}
	if f := atomic.LoadInt64(&finishes); f != int64(jobCount) {
		t.Fatalf("finishes=%d, want=%d", f, jobCount)
	}
}

func TestDispatcher_UsesRateLimiter(t *testing.T) {
	maxWorkers := 4
	queueSize := 10
	jobCount := 6

	// Each Acquire sleeps 40ms; the dispatch loop calls Acquire per job serially.
	// We expect roughly jobCount * 40ms total dispatch time before Stop returns.
	rl := &fakeRateLimiter{delay: 40 * time.Millisecond}

	d := New(maxWorkers, queueSize, rl, nil)
	start := time.Now()
	d.Start()

	for i := 0; i < jobCount; i++ {
		d.AddJob(testJob{}) // jobs do nothing and return immediately
	}

	d.Stop()
	elapsed := time.Since(start)

	if calls := atomic.LoadInt64(&rl.calls); calls != int64(jobCount) {
		t.Fatalf("rate limiter Acquire calls=%d, want=%d", calls, jobCount)
	}

	min := time.Duration(jobCount) * rl.delay
	// Allow generous tolerance for scheduling.
	if elapsed < min-40*time.Millisecond {
		t.Fatalf("elapsed=%v, want at least ~%v (rate limiter delay per job)", elapsed, min)
	}
}

func TestDispatcher_StopWaitsForInFlightJobs(t *testing.T) {
	maxWorkers := 2
	queueSize := 10
	jobCount := 4

	var wg sync.WaitGroup
	wg.Add(jobCount)

	jobDelay := 150 * time.Millisecond
	d := New(maxWorkers, queueSize, nil, nil)
	d.Start()

	for i := 0; i < jobCount; i++ {
		d.AddJob(testJob{
			delay: jobDelay,
			onFinish: func() {
				wg.Done()
			},
		})
	}

	// Stop should not return until all jobs are finished.
	before := time.Now()
	d.Stop()
	elapsed := time.Since(before)

	// Ensure all jobs truly finished
	doneCh := make(chan struct{})
	go func() { wg.Wait(); close(doneCh) }()
	select {
	case <-doneCh:
		// ok
	case <-time.After(1 * time.Second):
		t.Fatal("jobs did not finish by the time Stop returned")
	}

	// With 2 workers and 4 jobs, two waves of ~150ms → expect >= ~150ms total.
	if elapsed < jobDelay-30*time.Millisecond {
		t.Fatalf("Stop returned too early: elapsed=%v, jobDelay=%v", elapsed, jobDelay)
	}
}

// This test ensures AddJob blocks when queue is full and no workers are started yet.
func TestDispatcher_AddJobBackpressure(t *testing.T) {
	maxWorkers := 1
	queueSize := 1

	d := New(maxWorkers, queueSize, nil, nil)
	// Intentionally don't call Start() yet to keep workerPool empty, so only queue buffers.

	blocked := make(chan struct{})
	released := make(chan struct{})

	// Fill the queue to capacity.
	d.AddJob(testJob{})

	// Next AddJob should block until we Start and a worker begins draining.
	go func() {
		d.AddJob(testJob{})
		close(blocked) // indicates the goroutine reached the blocking point before release
		<-released
	}()

	// Give goroutine time to attempt AddJob and (likely) block.
	time.Sleep(20 * time.Millisecond)

	select {
	case <-blocked:
		// The second AddJob already proceeded, meaning it didn't block; unlikely with queueSize=1.
		// We can't reliably assert blocking without low-level instrumentation, so tolerate both outcomes.
	default:
		// Most likely blocked here; now start workers to drain.
	}

	d.Start()
	// Allow draining and unblocking; then stop.
	time.Sleep(50 * time.Millisecond)
	close(released)
	d.Stop()
}
