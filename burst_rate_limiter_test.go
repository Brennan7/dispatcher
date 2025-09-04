package dispatcher

import (
	"sync"
	"testing"
	"time"
)

// TestBurstRateLimiter_InitialBurstAllowsUpToMaxTokens verifies that an initial burst up to maxTokens proceeds immediately.
func TestBurstRateLimiter_InitialBurstAllowsUpToMaxTokens(t *testing.T) {
	rateLimit := 10 // tokens per second and bucket size
	limiter := NewBurstRateLimiter(rateLimit)

	start := time.Now()
	for i := 0; i < rateLimit; i++ {
		limiter.Acquire()
	}
	elapsed := time.Since(start)

	// The initial burst should be effectively immediate (allow some slack for scheduling).
	if elapsed > 100*time.Millisecond {
		t.Fatalf("initial burst of %d took %v; expected near-immediate (<100ms)", rateLimit, elapsed)
	}
}

// TestBurstRateLimiter_BlocksAfterExhaustionUntilRefill ensures that once the bucket is empty,
// the next Acquire blocks until a refill occurs. The exact wait depends on the ticker phase.
func TestBurstRateLimiter_BlocksAfterExhaustionUntilRefill(t *testing.T) {
	rateLimit := 8
	limiter := NewBurstRateLimiter(rateLimit)

	// Exhaust the bucket quickly.
	for i := 0; i < rateLimit; i++ {
		limiter.Acquire()
	}

	start := time.Now()
	limiter.Acquire() // should wait until tokens are refilled (ticker-phase dependent)
	elapsed := time.Since(start)

	// Expect a noticeable block; allow wide tolerance due to ticker phase and scheduler variability.
	if elapsed < 200*time.Millisecond || elapsed > 2*time.Second {
		t.Fatalf("Acquire after exhaustion blocked for %v; expected substantial delay (~0.2s–2s)", elapsed)
	}
}

// TestBurstRateLimiter_AverageRateOverWindow checks that total acquisitions over a window
// are bounded by initial tokens + refill rate * duration.
func TestBurstRateLimiter_AverageRateOverWindow(t *testing.T) {
	rateLimit := 5
	limiter := NewBurstRateLimiter(rateLimit)

	window := 2 * time.Second
	deadline := time.Now().Add(window)

	count := 0
	for time.Now().Before(deadline) {
		limiter.Acquire()
		count++
	}

	// Theoretical max over 2s = initial bucket (5) + rate*2s (10) = 15.
	// We should also meet at least roughly rate*window (10); allow 1 slack.
	min := rateLimit * int(window/time.Second)           // 10
	max := rateLimit + rateLimit*int(window/time.Second) // 15

	if count < min-1 || count > max+1 {
		t.Fatalf("acquired %d tokens in %v; expected between ~%d and ~%d (±1)", count, window, min, max)
	}
}

// TestBurstRateLimiter_DoesNotOverfill ensures the bucket cap is enforced after idle time.
func TestBurstRateLimiter_DoesNotOverfill(t *testing.T) {
	rateLimit := 6
	limiter := NewBurstRateLimiter(rateLimit)

	// Let it idle for >1s so the bucket would be full, but not more than maxTokens.
	time.Sleep(1200 * time.Millisecond)

	// Consume exactly maxTokens; should be near-immediate.
	start := time.Now()
	for i := 0; i < rateLimit; i++ {
		limiter.Acquire()
	}
	elapsedBurst := time.Since(start)
	if elapsedBurst > 100*time.Millisecond {
		t.Fatalf("consuming %d tokens from full bucket took %v; expected near-immediate (<100ms)", rateLimit, elapsedBurst)
	}

	// One more acquire should block until a refill occurs.
	// Because the refill ticker isn't phase-aligned with our sleep, we assert a broad window.
	startBlock := time.Now()
	limiter.Acquire()
	blockElapsed := time.Since(startBlock)

	// Expect a significant block (bucket is empty), but allow wide variance due to ticker phase and scheduler.
	if blockElapsed < 500*time.Millisecond || blockElapsed > 2*time.Second {
		t.Fatalf("Acquire beyond cap blocked for %v; expected substantial delay (~0.5s–2s)", blockElapsed)
	}
}

// TestBurstRateLimiter_ConcurrentAcquire verifies basic concurrency safety and aggregate timing.
// With N=10/s and total=20 acquires, we expect ~1s wall time: first 10 immediately, next 10 after refill.
func TestBurstRateLimiter_ConcurrentAcquire(t *testing.T) {
	rateLimit := 10
	total := 20
	limiter := NewBurstRateLimiter(rateLimit)

	var wg sync.WaitGroup
	wg.Add(total)

	start := time.Now()
	for i := 0; i < total; i++ {
		go func() {
			defer wg.Done()
			limiter.Acquire()
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	// Expect roughly 1s; allow broad tolerance for scheduler and platform variance.
	if elapsed < 700*time.Millisecond || elapsed > 2500*time.Millisecond {
		t.Fatalf("concurrent %d acquires took %v; expected around ~1s (0.7–2.5s tolerance)", total, elapsed)
	}
}
