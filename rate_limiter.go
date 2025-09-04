package dispatcher

// RateLimiter defines an interface for different rate-limiting strategies
type RateLimiter interface {
	Acquire()
}
