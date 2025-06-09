package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"voice-chat-app/errors"
	"voice-chat-app/handlers"
	"voice-chat-app/middleware"
	"voice-chat-app/models"
	"voice-chat-app/utils"
)

func main() {
	// Load configuration
	config := utils.LoadConfig()

	// Initialize logger
	utils.InitLogger(config)
	ctx := context.Background()

	utils.Info(ctx, "Starting voice chat server", map[string]interface{}{
		"port":               config.Port,
		"environment":        config.Environment,
		"log_level":          config.LogLevel,
		"stun_servers":       config.STUNServers,
		"turn_servers_count": len(config.TURNServers),
		"allowed_origins":    config.AllowedOrigins,
		"max_connections":    config.MaxConnections,
		"http_rate_limit":    config.HTTPRateLimitPerMinute,
		"ws_rate_limit":      config.WSRateLimitPerMinute,
	})

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(
		config.HTTPRateLimitPerMinute,
		config.WSRateLimitPerMinute,
		config.MaxWSConnPerIP,
	)

	// Initialize CORS configuration
	corsConfig := middleware.NewCORSConfig(config.AllowedOrigins)

	// Initialize user pool
	userPool := models.NewUserPool()

	// Initialize signaling server with enhanced configuration
	signalingServer := &handlers.SignalingServer{
		UserPool:    userPool,
		RateLimiter: rateLimiter,
		STUNServers: config.STUNServers,
		TURNServers: convertTURNServers(config.TURNServers),
	}

	// Create HTTP mux
	mux := http.NewServeMux()

	// Setup routes
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Log WebSocket connection attempts
		clientIP := r.Header.Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP = r.RemoteAddr
		}
		utils.Info(ctx, "WebSocket connection attempt", map[string]interface{}{
			"client_ip":  clientIP,
			"user_agent": r.Header.Get("User-Agent"),
			"origin":     r.Header.Get("Origin"),
			"path":       r.URL.Path,
		})

		signalingServer.HandleWebSocket(w, r)
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "healthy",
			"time":        time.Now().Format(time.RFC3339),
			"environment": config.Environment,
			"version":     "1.0.0",
		})
	})

	// Enhanced stats endpoint with rate limiting info
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		stats := signalingServer.GetStats()
		stats["rate_limiter"] = rateLimiter.GetStats()
		stats["jwt_blacklist"] = utils.GetBlacklistStats()

		json.NewEncoder(w).Encode(stats)
	})

	// ICE servers endpoint for mobile clients
	mux.HandleFunc("/ice-servers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		iceServers := signalingServer.GetICEServers()
		json.NewEncoder(w).Encode(iceServers)
	})

	// Apply middleware stack (order matters!)
	handler := middleware.Chain(
		mux,
		errors.ErrorHandler,       // Error handling (outermost)
		corsConfig.CORS,           // CORS handling
		rateLimiter.HTTPRateLimit, // Rate limiting
		utils.LoggerMiddleware,    // Request logging (innermost)
	)

	// Create server with configuration
	server := &http.Server{
		Addr:           ":" + config.Port,
		Handler:        handler,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		IdleTimeout:    config.IdleTimeout,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Start server in a goroutine
	go func() {
		utils.Info(ctx, "Voice chat server starting", map[string]interface{}{
			"address": server.Addr,
		})

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Fatal(ctx, "Server failed to start", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	utils.Info(ctx, "Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown user pool
	userPool.Shutdown()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		utils.Fatal(ctx, "Server forced to shutdown", err)
	}

	utils.Info(ctx, "Server exited gracefully")
}

// convertTURNServers converts config TURN servers to handler TURN servers
func convertTURNServers(configServers []utils.TURNServerConfig) []handlers.TURNServer {
	var turnServers []handlers.TURNServer
	for _, server := range configServers {
		turnServers = append(turnServers, handlers.TURNServer{
			URL:        server.URL,
			Username:   server.Username,
			Credential: server.Credential,
		})
	}
	return turnServers
}
