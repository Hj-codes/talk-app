package middleware

import "net/http"

// Middleware represents a middleware function
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares to a handler in the order they are provided
// The first middleware in the list will be the outermost middleware
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	// Apply middlewares in reverse order so the first one becomes outermost
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
