package middleware

import (
	"fmt"
	"net/http"
	"strings"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// NewCORSConfig creates a new CORS configuration
func NewCORSConfig(allowedOrigins []string) *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Requested-With",
		},
		ExposedHeaders: []string{
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		AllowCredentials: false, // Set to true only if needed
		MaxAge:           3600,  // 1 hour
	}
}

// CORS middleware that applies CORS headers based on configuration
func (c *CORSConfig) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		if origin != "" && c.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(c.AllowedOrigins) == 1 && c.AllowedOrigins[0] == "*" {
			// Only allow wildcard in development
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		// Set other CORS headers
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(c.AllowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(c.AllowedHeaders, ", "))

		if len(c.ExposedHeaders) > 0 {
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(c.ExposedHeaders, ", "))
		}

		if c.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if c.MaxAge > 0 {
			w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", c.MaxAge))
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isOriginAllowed checks if the given origin is in the allowed list
func (c *CORSConfig) isOriginAllowed(origin string) bool {
	for _, allowedOrigin := range c.AllowedOrigins {
		if allowedOrigin == "*" {
			return true
		}
		if allowedOrigin == origin {
			return true
		}
		// Support wildcard subdomains like *.example.com
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := allowedOrigin[2:]
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true
			}
		}
	}
	return false
}

// ValidateOrigin performs additional security checks on the origin
func (c *CORSConfig) ValidateOrigin(origin string) bool {
	if origin == "" {
		return false
	}

	// Check for dangerous patterns
	dangerousPatterns := []string{
		"javascript:",
		"data:",
		"vbscript:",
		"file:",
		"about:",
	}

	lowerOrigin := strings.ToLower(origin)
	for _, pattern := range dangerousPatterns {
		if strings.HasPrefix(lowerOrigin, pattern) {
			return false
		}
	}

	// Must start with http:// or https://
	if !strings.HasPrefix(lowerOrigin, "http://") && !strings.HasPrefix(lowerOrigin, "https://") {
		return false
	}

	return c.isOriginAllowed(origin)
}
