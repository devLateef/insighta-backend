package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// slidingWindow tracks request counts in a 1-minute sliding window per key.
type slidingWindow struct {
	mu       sync.Mutex
	windows  map[string][]time.Time
	maxReqs  int
	interval time.Duration
}

func newSlidingWindow(maxReqs int, interval time.Duration) *slidingWindow {
	sw := &slidingWindow{
		windows:  make(map[string][]time.Time),
		maxReqs:  maxReqs,
		interval: interval,
	}
	go sw.cleanup()
	return sw
}

func (sw *slidingWindow) allow(key string) bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.interval)

	// Remove timestamps outside the window
	times := sw.windows[key]
	valid := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= sw.maxReqs {
		sw.windows[key] = valid
		return false
	}

	sw.windows[key] = append(valid, now)
	return true
}

func (sw *slidingWindow) cleanup() {
	for {
		time.Sleep(2 * time.Minute)
		sw.mu.Lock()
		cutoff := time.Now().Add(-sw.interval)
		for key, times := range sw.windows {
			valid := times[:0]
			for _, t := range times {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(sw.windows, key)
			} else {
				sw.windows[key] = valid
			}
		}
		sw.mu.Unlock()
	}
}

// Auth endpoints: 10 req/min per IP (strict sliding window)
var authLimiter = newSlidingWindow(10, time.Minute)

// API endpoints: 60 req/min per user
var apiLimiter = newSlidingWindow(60, time.Minute)

// RateLimitAuth applies 10 req/min limit for auth endpoints (per IP).
func RateLimitAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if !authLimiter.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":  "error",
				"message": "too many requests, please slow down",
			})
			return
		}
		c.Next()
	}
}

// RateLimitAPI applies 60 req/min limit for API endpoints (per user ID, falls back to IP).
func RateLimitAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetString("user_id")
		if key == "" {
			key = c.ClientIP()
		}
		if !apiLimiter.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":  "error",
				"message": "too many requests, please slow down",
			})
			return
		}
		c.Next()
	}
}
