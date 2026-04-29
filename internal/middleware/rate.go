package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiterStore struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     rate.Limit
	burst    int
}

func newStore(r rate.Limit, burst int) *rateLimiterStore {
	s := &rateLimiterStore{
		visitors: make(map[string]*visitor),
		rate:     r,
		burst:    burst,
	}
	go s.cleanup()
	return s
}

func (s *rateLimiterStore) get(key string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, exists := s.visitors[key]
	if !exists {
		lim := rate.NewLimiter(s.rate, s.burst)
		s.visitors[key] = &visitor{limiter: lim, lastSeen: time.Now()}
		return lim
	}
	v.lastSeen = time.Now()
	return v.limiter
}

func (s *rateLimiterStore) cleanup() {
	for {
		time.Sleep(time.Minute)
		s.mu.Lock()
		for key, v := range s.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(s.visitors, key)
			}
		}
		s.mu.Unlock()
	}
}

// Auth endpoints: 10 req/min
var authStore = newStore(rate.Every(6*time.Second), 10)

// API endpoints: 60 req/min per user
var apiStore = newStore(rate.Every(time.Second), 60)

// RateLimitAuth applies 10 req/min limit for auth endpoints (per IP).
func RateLimitAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if !authStore.get(key).Allow() {
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
		if !apiStore.get(key).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":  "error",
				"message": "too many requests, please slow down",
			})
			return
		}
		c.Next()
	}
}
