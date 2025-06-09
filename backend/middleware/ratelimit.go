package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter manages rate limiting for different types of requests
type RateLimiter struct {
	httpLimiters    map[string]*rate.Limiter
	wsLimiters      map[string]*rate.Limiter
	wsConnections   map[string]int
	mutex           sync.RWMutex
	httpRate        rate.Limit
	httpBurst       int
	wsRate          rate.Limit
	wsBurst         int
	maxWSConnPerIP  int
	cleanupInterval time.Duration
}

// NewRateLimiter creates a new rate limiter with specified rates
func NewRateLimiter(httpRequestsPerMinute, wsRequestsPerMinute, maxWSConnPerIP int) *RateLimiter {
	rl := &RateLimiter{
		httpLimiters:    make(map[string]*rate.Limiter),
		wsLimiters:      make(map[string]*rate.Limiter),
		wsConnections:   make(map[string]int),
		httpRate:        rate.Limit(httpRequestsPerMinute) / 60, // per second
		httpBurst:       httpRequestsPerMinute / 4,              // allow burst of 1/4 of per-minute rate
		wsRate:          rate.Limit(wsRequestsPerMinute) / 60,   // per second
		wsBurst:         wsRequestsPerMinute / 4,                // allow burst of 1/4 of per-minute rate
		maxWSConnPerIP:  maxWSConnPerIP,
		cleanupInterval: 5 * time.Minute,
	}

	// Start cleanup goroutine
	go rl.cleanupExpiredLimiters()

	return rl
}

// HTTPRateLimit middleware for HTTP requests
func (rl *RateLimiter) HTTPRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		if !rl.allowHTTPRequest(ip) {
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%.0f", float64(rl.httpRate*60)))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// Add rate limit headers for successful requests
		limiter := rl.getHTTPLimiter(ip)
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%.0f", float64(rl.httpRate*60)))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%.0f", float64(limiter.Tokens())))

		next.ServeHTTP(w, r)
	})
}

// CheckWebSocketRateLimit checks if a WebSocket request should be allowed
func (rl *RateLimiter) CheckWebSocketRateLimit(ip string) bool {
	return rl.allowWSRequest(ip)
}

// CheckWebSocketConnection checks if a new WebSocket connection should be allowed
func (rl *RateLimiter) CheckWebSocketConnection(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	currentConnections := rl.wsConnections[ip]
	if currentConnections >= rl.maxWSConnPerIP {
		return false
	}

	rl.wsConnections[ip] = currentConnections + 1
	return true
}

// ReleaseWebSocketConnection releases a WebSocket connection for an IP
func (rl *RateLimiter) ReleaseWebSocketConnection(ip string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if count, exists := rl.wsConnections[ip]; exists && count > 0 {
		rl.wsConnections[ip] = count - 1
		if rl.wsConnections[ip] == 0 {
			delete(rl.wsConnections, ip)
		}
	}
}

// allowHTTPRequest checks if an HTTP request should be allowed
func (rl *RateLimiter) allowHTTPRequest(ip string) bool {
	limiter := rl.getHTTPLimiter(ip)
	return limiter.Allow()
}

// allowWSRequest checks if a WebSocket message should be allowed
func (rl *RateLimiter) allowWSRequest(ip string) bool {
	limiter := rl.getWSLimiter(ip)
	return limiter.Allow()
}

// getHTTPLimiter gets or creates an HTTP rate limiter for an IP
func (rl *RateLimiter) getHTTPLimiter(ip string) *rate.Limiter {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	limiter, exists := rl.httpLimiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.httpRate, rl.httpBurst)
		rl.httpLimiters[ip] = limiter
	}

	return limiter
}

// getWSLimiter gets or creates a WebSocket rate limiter for an IP
func (rl *RateLimiter) getWSLimiter(ip string) *rate.Limiter {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	limiter, exists := rl.wsLimiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.wsRate, rl.wsBurst)
		rl.wsLimiters[ip] = limiter
	}

	return limiter
}

// cleanupExpiredLimiters removes unused rate limiters periodically
func (rl *RateLimiter) cleanupExpiredLimiters() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()

		// Remove limiters that haven't been used recently
		for ip, limiter := range rl.httpLimiters {
			// If limiter is at full capacity, it hasn't been used recently
			if limiter.Tokens() >= float64(rl.httpBurst) {
				delete(rl.httpLimiters, ip)
			}
		}

		for ip, limiter := range rl.wsLimiters {
			// If limiter is at full capacity, it hasn't been used recently
			if limiter.Tokens() >= float64(rl.wsBurst) {
				delete(rl.wsLimiters, ip)
			}
		}

		rl.mutex.Unlock()
	}
}

// getClientIP extracts the real client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for reverse proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP from the list
		if ip := parseFirstIP(xff); ip != "" {
			return ip
		}
	}

	// Check X-Real-IP header (for nginx)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if ip := net.ParseIP(xri); ip != nil {
			return xri
		}
	}

	// Fall back to remote address
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// parseFirstIP parses the first valid IP from a comma-separated list
func parseFirstIP(ips string) string {
	for _, ip := range []string{ips} {
		// Split by comma and take first
		if idx := strings.Index(ip, ","); idx != -1 {
			ip = ip[:idx]
		}
		ip = strings.TrimSpace(ip)
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	return ""
}

// GetStats returns current rate limiting statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	stats := map[string]interface{}{
		"active_http_limiters": len(rl.httpLimiters),
		"active_ws_limiters":   len(rl.wsLimiters),
		"active_ws_connections": func() int {
			total := 0
			for _, count := range rl.wsConnections {
				total += count
			}
			return total
		}(),
		"unique_ips_with_ws_connections": len(rl.wsConnections),
		"http_rate_per_minute":           float64(rl.httpRate * 60),
		"ws_rate_per_minute":             float64(rl.wsRate * 60),
		"max_ws_connections_per_ip":      rl.maxWSConnPerIP,
	}

	return stats
}
