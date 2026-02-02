package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMiddleware(t *testing.T) {
	m := newMiddleware("test-key", 10, 60*time.Second)
	assert.NotNil(t, m)
	assert.Equal(t, "test-key", m.apiKey)
	assert.Equal(t, 10, m.rateLimiter.rate)
	assert.Equal(t, 60*time.Second, m.rateLimiter.window)
}

func TestMiddlewareAuth_NoKey(t *testing.T) {
	m := newMiddleware("", 0, time.Second)
	handler := m.auth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestMiddlewareAuth_WithValidKey(t *testing.T) {
	m := newMiddleware("secret-key", 0, time.Second)
	handler := m.auth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", "secret-key")
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestMiddlewareAuth_WithInvalidKey(t *testing.T) {
	m := newMiddleware("secret-key", 0, time.Second)
	handler := m.auth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Unauthorized")
}

func TestMiddlewareAuth_NoHeader(t *testing.T) {
	m := newMiddleware("secret-key", 0, time.Second)
	handler := m.auth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddlewareRateLimit_Disabled(t *testing.T) {
	m := newMiddleware("", 0, time.Second)
	handler := m.rateLimit(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for i := 0; i < 200; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i)
	}
}

func TestMiddlewareRateLimit_Enabled(t *testing.T) {
	rate := 5
	window := 1 * time.Second
	m := newMiddleware("", rate, window)
	handler := m.rateLimit(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	for i := 0; i < rate; i++ {
		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i)
	}

	w := httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Rate limit exceeded")
}

func TestMiddlewareRateLimit_DifferentIPs(t *testing.T) {
	rate := 3
	window := 1 * time.Second
	m := newMiddleware("", rate, window)
	handler := m.rateLimit(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ips := []string{"192.168.1.1:1234", "192.168.1.2:1234", "192.168.1.3:1234"}

	for _, ip := range ips {
		for i := 0; i < rate; i++ {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = ip
			w := httptest.NewRecorder()
			handler(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}
	}
}

func TestMiddlewareRateLimit_WindowReset(t *testing.T) {
	rate := 2
	window := 100 * time.Millisecond
	m := newMiddleware("", rate, window)
	handler := m.rateLimit(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	for i := 0; i < rate; i++ {
		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	w := httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	time.Sleep(window + 50*time.Millisecond)

	w = httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Should succeed after window reset")
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := &rateLimiter{
		clients: make(map[string]*clientLimiter),
		rate:    10,
		window:  1 * time.Second,
	}

	rl.clients["192.168.1.1"] = &clientLimiter{
		tokens: 5,
		last:   time.Now(),
	}

	rl.clients["192.168.1.2"] = &clientLimiter{
		tokens: 5,
		last:   time.Now().Add(-10 * time.Second),
	}

	assert.Equal(t, 2, len(rl.clients))

	rl.cleanup()

	assert.Equal(t, 1, len(rl.clients))
	assert.Contains(t, rl.clients, "192.168.1.1")
	assert.NotContains(t, rl.clients, "192.168.1.2")
}

func TestClientLimiter_ConcurrentAccess(t *testing.T) {
	rate := 100
	window := 1 * time.Second
	m := newMiddleware("", rate, window)
	handler := m.rateLimit(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	done := make(chan bool)
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	for i := 0; i < 50; i++ {
		go func() {
			w := httptest.NewRecorder()
			handler(w, req)
			done <- true
		}()
	}

	for i := 0; i < 50; i++ {
		<-done
	}
}
