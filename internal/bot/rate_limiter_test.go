package bot

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiterAllowsUpToLimit(t *testing.T) {
	rl := NewRateLimiter()
	for i := range rateLimitMaxCommands {
		require.True(t, rl.Allow("user-1"), "request %d should be allowed", i+1)
	}
	assert.False(t, rl.Allow("user-1"), "request beyond limit should be denied")
}

func TestRateLimiterIsolatesUsers(t *testing.T) {
	rl := NewRateLimiter()
	for range rateLimitMaxCommands {
		rl.Allow("user-1")
	}
	assert.False(t, rl.Allow("user-1"))
	assert.True(t, rl.Allow("user-2"), "different user should not be affected")
}

func TestRateLimiterResetsAfterWindow(t *testing.T) {
	rl := NewRateLimiter()

	// Fill up the limit by backdating timestamps
	rl.mu.Lock()
	past := time.Now().Add(-rateLimitWindow - time.Second)
	for range rateLimitMaxCommands {
		rl.requests["user-1"] = append(rl.requests["user-1"], past)
	}
	rl.mu.Unlock()

	assert.True(t, rl.Allow("user-1"), "should allow after old entries expire")
}

func TestRateLimiterPrunesOldEntries(t *testing.T) {
	rl := NewRateLimiter()

	rl.mu.Lock()
	old := time.Now().Add(-rateLimitWindow - time.Second)
	for range 3 {
		rl.requests["user-1"] = append(rl.requests["user-1"], old)
	}
	rl.mu.Unlock()

	// Should prune the 3 old entries and allow new ones
	for i := range rateLimitMaxCommands {
		require.True(t, rl.Allow("user-1"), "request %d should be allowed after pruning", i+1)
	}
	assert.False(t, rl.Allow("user-1"))
}

func TestRateLimiterConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter()
	var wg sync.WaitGroup
	allowed := make([]int, 10)

	for i := range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			userID := fmt.Sprintf("user-%d", i)
			for range rateLimitMaxCommands + 2 {
				if rl.Allow(userID) {
					allowed[i]++
				}
			}
		}()
	}
	wg.Wait()

	for i, count := range allowed {
		assert.Equal(t, rateLimitMaxCommands, count, "user-%d should have exactly %d allowed requests", i, rateLimitMaxCommands)
	}
}
