package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type bucket struct {
	tokens   float64
	lastSeen time.Time
}

// RateLimit returns a token-bucket rate limiter middleware keyed on a value
// extracted by keyFn (e.g. the X-API-Key header).
//
// rate  – tokens added per second
// burst – maximum tokens (= max requests in a burst)
func RateLimit(rate float64, burst float64, keyFn func(*http.Request) string) func(http.Handler) http.Handler {
	var mu sync.Mutex
	buckets := make(map[string]*bucket)

	// Periodically clean up stale buckets
	go func() {
		for range time.Tick(5 * time.Minute) {
			mu.Lock()
			for k, b := range buckets {
				if time.Since(b.lastSeen) > 10*time.Minute {
					delete(buckets, k)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFn(r)
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			mu.Lock()
			b, ok := buckets[key]
			if !ok {
				b = &bucket{tokens: burst}
				buckets[key] = b
			}

			now := time.Now()
			elapsed := now.Sub(b.lastSeen).Seconds()
			b.lastSeen = now
			b.tokens = min(burst, b.tokens+elapsed*rate)

			if b.tokens < 1 {
				mu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "rate limit exceeded"})
				return
			}

			b.tokens--
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
