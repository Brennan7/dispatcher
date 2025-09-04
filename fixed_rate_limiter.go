package dispatcher

import "time"

// FixedRateLimiter enforces a strict rate limit by evenly spacing job dispatches
type FixedRateLimiter struct {
	ticker *time.Ticker
}

// NewFixedRateLimiter initializes a ticker that dispatches at a fixed rate
func NewFixedRateLimiter(rateLimit int) *FixedRateLimiter {
	return &FixedRateLimiter{
		ticker: time.NewTicker(time.Second / time.Duration(rateLimit)),
	}
}

// Acquire waits for the next tick before allowing a job to be processed
func (t *FixedRateLimiter) Acquire() {
	<-t.ticker.C // Wait for the next tick
}
