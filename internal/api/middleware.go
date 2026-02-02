package api

import (
	"net/http"
	"sync"
	"time"
)

type middleware struct {
	apiKey      string
	rateLimiter *rateLimiter
}

type rateLimiter struct {
	clients map[string]*clientLimiter
	mu      sync.RWMutex
	rate    int
	window  time.Duration
}

type clientLimiter struct {
	tokens int
	last   time.Time
	mu     sync.Mutex
}

func newMiddleware(apiKey string, rateLimit int, rateWindow time.Duration) *middleware {
	return &middleware{
		apiKey: apiKey,
		rateLimiter: &rateLimiter{
			clients: make(map[string]*clientLimiter),
			rate:    rateLimit,
			window:  rateWindow,
		},
	}
}

func (m *middleware) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if m.apiKey == "" {
			next.ServeHTTP(w, r)
			return
		}

		key := r.Header.Get("X-API-Key")
		if key != m.apiKey {
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (m *middleware) rateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if m.rateLimiter.rate == 0 {
			next.ServeHTTP(w, r)
			return
		}

		clientIP := r.RemoteAddr
		if !m.rateLimiter.allow(clientIP) {
			http.Error(w, `{"error":"Rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cl, exists := rl.clients[ip]
	if !exists {
		cl = &clientLimiter{
			tokens: rl.rate - 1,
			last:   time.Now(),
		}
		rl.clients[ip] = cl
		return true
	}

	cl.mu.Lock()
	defer cl.mu.Unlock()

	elapsed := time.Since(cl.last)
	if elapsed >= rl.window {
		cl.tokens = rl.rate - 1
		cl.last = time.Now()
		return true
	}

	if cl.tokens > 0 {
		cl.tokens--
		return true
	}

	return false
}

func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for ip, cl := range rl.clients {
		cl.mu.Lock()
		if time.Since(cl.last) > rl.window*5 {
			delete(rl.clients, ip)
		}
		cl.mu.Unlock()
	}
}
