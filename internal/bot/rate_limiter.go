package bot

import (
	"sync"
	"time"
)

const (
	rateLimitMaxCommands = 5
	rateLimitWindow      = 60 * time.Second
)

type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
	}
}

func (r *RateLimiter) Allow(userID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rateLimitWindow)

	timestamps := r.requests[userID]
	pruned := timestamps[:0]
	for _, t := range timestamps {
		if t.After(cutoff) {
			pruned = append(pruned, t)
		}
	}

	if len(pruned) >= rateLimitMaxCommands {
		r.requests[userID] = pruned
		return false
	}

	r.requests[userID] = append(pruned, now)
	return true
}
