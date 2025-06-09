package utils

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// ContextKey type for context keys
type ContextKey string

const (
	CorrelationIDKey ContextKey = "correlation_id"
	UserIDKey        ContextKey = "user_id"
	SessionIDKey     ContextKey = "session_id"
	IPAddressKey     ContextKey = "ip_address"
)

var logger *logrus.Logger

// InitLogger initializes the global logger with configuration
func InitLogger(config *Config) {
	logger = logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set output
	logger.SetOutput(os.Stdout)

	// Set formatter based on environment
	if config.IsProduction() {
		// JSON formatter for production (better for log aggregation)
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "function",
				logrus.FieldKeyFile:  "file",
			},
		})
	} else {
		// Text formatter for development (better readability)
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		})
	}

	// Add caller information in development
	if config.IsDevelopment() {
		logger.SetReportCaller(true)
	}
}

// GetLogger returns the global logger instance
func GetLogger() *logrus.Logger {
	if logger == nil {
		// Fallback initialization
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}
	return logger
}

// NewLoggerEntry creates a new logger entry with context
func NewLoggerEntry(ctx context.Context) *logrus.Entry {
	entry := GetLogger().WithFields(logrus.Fields{})

	// Add correlation ID if present
	if correlationID := GetCorrelationID(ctx); correlationID != "" {
		entry = entry.WithField("correlation_id", correlationID)
	}

	// Add user ID if present
	if userID := GetUserID(ctx); userID != "" {
		entry = entry.WithField("user_id", userID)
	}

	// Add session ID if present
	if sessionID := GetSessionID(ctx); sessionID != "" {
		entry = entry.WithField("session_id", sessionID)
	}

	// Add IP address if present
	if ipAddress := GetIPAddress(ctx); ipAddress != "" {
		entry = entry.WithField("ip_address", ipAddress)
	}

	return entry
}

// Context helpers

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// GetCorrelationID retrieves the correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return correlationID
	}
	return ""
}

// WithUserID adds a user ID to the context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID retrieves the user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// WithSessionID adds a session ID to the context
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, SessionIDKey, sessionID)
}

// GetSessionID retrieves the session ID from context
func GetSessionID(ctx context.Context) string {
	if sessionID, ok := ctx.Value(SessionIDKey).(string); ok {
		return sessionID
	}
	return ""
}

// WithIPAddress adds an IP address to the context
func WithIPAddress(ctx context.Context, ipAddress string) context.Context {
	return context.WithValue(ctx, IPAddressKey, ipAddress)
}

// GetIPAddress retrieves the IP address from context
func GetIPAddress(ctx context.Context) string {
	if ipAddress, ok := ctx.Value(IPAddressKey).(string); ok {
		return ipAddress
	}
	return ""
}

// Convenience logging functions

// Debug logs a debug message with context
func Debug(ctx context.Context, msg string, fields ...logrus.Fields) {
	entry := NewLoggerEntry(ctx)
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Debug(msg)
}

// Info logs an info message with context
func Info(ctx context.Context, msg string, fields ...logrus.Fields) {
	entry := NewLoggerEntry(ctx)
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Info(msg)
}

// Warn logs a warning message with context
func Warn(ctx context.Context, msg string, fields ...logrus.Fields) {
	entry := NewLoggerEntry(ctx)
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Warn(msg)
}

// Error logs an error message with context
func Error(ctx context.Context, msg string, err error, fields ...logrus.Fields) {
	entry := NewLoggerEntry(ctx)
	if err != nil {
		entry = entry.WithError(err)
	}
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Error(msg)
}

// Fatal logs a fatal message with context and exits
func Fatal(ctx context.Context, msg string, err error, fields ...logrus.Fields) {
	entry := NewLoggerEntry(ctx)
	if err != nil {
		entry = entry.WithError(err)
	}
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Fatal(msg)
}

// LoggerMiddleware creates a middleware that adds correlation ID to requests
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate correlation ID
		correlationID := GenerateUUID()

		// Add to request context
		ctx := WithCorrelationID(r.Context(), correlationID)
		ctx = WithIPAddress(ctx, getClientIP(r))

		// Add correlation ID to response headers
		w.Header().Set("X-Correlation-ID", correlationID)

		// Update request with new context
		r = r.WithContext(ctx)

		// Log request
		Info(ctx, "HTTP request", logrus.Fields{
			"method":     r.Method,
			"path":       r.URL.Path,
			"user_agent": r.Header.Get("User-Agent"),
			"referer":    r.Header.Get("Referer"),
		})

		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts client IP from request (reused from middleware)
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return parseFirstIP(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to remote address
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// parseFirstIP parses the first IP from comma-separated list
func parseFirstIP(ips string) string {
	if idx := strings.Index(ips, ","); idx != -1 {
		return strings.TrimSpace(ips[:idx])
	}
	return strings.TrimSpace(ips)
}

// SetLogOutput sets the output for the logger (useful for testing)
func SetLogOutput(output io.Writer) {
	if logger != nil {
		logger.SetOutput(output)
	}
}

// SetLogLevel sets the log level for the logger
func SetLogLevel(level string) error {
	if logger == nil {
		return nil
	}

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}

	logger.SetLevel(logLevel)
	return nil
}
