package dispatcher

import (
	"sync"
	"time"
)

// BurstRateLimiter allows bursts but enforces an overall processing limit
type BurstRateLimiter struct {
	maxTokens int
	tokens    int
	mu        sync.Mutex
}

// NewBurstRateLimiter initializes a token bucket rate limiter
func NewBurstRateLimiter(rateLimit int) *BurstRateLimiter {
	rl := &BurstRateLimiter{
		maxTokens: rateLimit,
		tokens:    rateLimit, // Start with a full bucket
	}

	// Refill tokens in a separate goroutine.
	go rl.refillTokens()
	return rl
}

// Acquire WaitForToken waits until a token is available before proceeding
func (rl *BurstRateLimiter) Acquire() {
	for {
		rl.mu.Lock()
		if rl.tokens > 0 {
			rl.tokens--
			rl.mu.Unlock()
			return
		}
		rl.mu.Unlock()
		time.Sleep(5 * time.Millisecond) // Wait for refills
	}
}

/*
refillTokens periodically replenishes tokens based on elapsed time. It calculates how many tokens should be refilled
in bulk based on the time since the last update. This ensures tokens never exceed `maxTokens`.
*/
func (rl *BurstRateLimiter) refillTokens() {
	ticker := time.NewTicker(time.Second) // Check every second
	defer ticker.Stop()

	lastRefill := time.Now()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(lastRefill).Seconds() // Calculate time elapsed since last refill

		// Calculate how many tokens to add (up to maxTokens)
		newTokens := int(elapsed * float64(rl.maxTokens))
		if newTokens > 0 {
			rl.tokens = min(rl.tokens+newTokens, rl.maxTokens)
			lastRefill = now
		}
		rl.mu.Unlock()
	}
}
