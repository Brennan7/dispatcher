package dispatcher

import (
	"testing"
	"time"
)

// TestFixedRateLimiter_WaitForToken verifies that Acquire correctly blocks for the expected time
func TestFixedRateLimiter_WaitForToken(t *testing.T) {
	rateLimit := 10 // 10 jobs per second (100ms per job)
	limiter := NewFixedRateLimiter(rateLimit)

	start := time.Now()
	limiter.Acquire()
	elapsed := time.Since(start)

	expectedDuration := time.Second / time.Duration(rateLimit)
	if elapsed < expectedDuration*9/10 || elapsed > expectedDuration*12/10 {
		t.Errorf("Acquire() blocked for %v, expected ~%v", elapsed, expectedDuration)
	}
}

// TestFixedRateLimiter_MultipleCalls ensures multiple calls respect the rate limit
func TestFixedRateLimiter_MultipleCalls(t *testing.T) {
	rateLimit := 5 // 5 jobs per second (200ms per job)
	limiter := NewFixedRateLimiter(rateLimit)

	start := time.Now()
	for i := 0; i < 3; i++ {
		limiter.Acquire()
	}

	elapsed := time.Since(start)
	expectedTotalDuration := 3 * (time.Second / time.Duration(rateLimit))

	if elapsed < expectedTotalDuration*9/10 || elapsed > expectedTotalDuration*12/10 {
		t.Errorf("3 calls to Acquire() took %v, expected ~%v", elapsed, expectedTotalDuration)
	}
}

// TestFixedRateLimiter_LowRateLimit ensures low rate limits enforce long wait times
func TestFixedRateLimiter_LowRateLimit(t *testing.T) {
	rateLimit := 1 // 1 job per second (1000ms per job)
	limiter := NewFixedRateLimiter(rateLimit)

	start := time.Now()
	limiter.Acquire()
	elapsed := time.Since(start)

	if elapsed < 900*time.Millisecond || elapsed > 1100*time.Millisecond {
		t.Errorf("Acquire() blocked for %v, expected ~1s", elapsed)
	}
}
