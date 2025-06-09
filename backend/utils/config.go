package utils

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"voice-chat-app/models"
)

type Config struct {
	// Server configuration
	Port        string
	Environment string
	LogLevel    string

	// Security configuration
	JWTSecret      []byte
	AllowedOrigins []string

	// Timeout configuration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	HeartbeatInterval time.Duration
	CleanupInterval   time.Duration
	ConnectionTimeout time.Duration
	WebSocketTimeout  time.Duration

	// Rate limiting configuration
	MaxConnections         int
	HTTPRateLimitPerMinute int
	WSRateLimitPerMinute   int
	MaxWSConnPerIP         int

	// WebRTC configuration
	STUNServers []string
	TURNServers []TURNServerConfig
}

type TURNServerConfig struct {
	URL        string
	Username   string
	Credential string
}

func LoadConfig() *Config {
	config := &Config{
		// Server settings
		Port:        getEnv(models.EnvPort, "8080"),
		Environment: getEnv(models.EnvEnvironment, models.EnvironmentDevelopment),
		LogLevel:    getEnv(models.EnvLogLevel, models.LogLevelInfo),

		// Security settings
		JWTSecret:      getJWTSecret(),
		AllowedOrigins: getAllowedOrigins(),

		// Timeout settings
		ReadTimeout:       getDurationEnv("READ_TIMEOUT", models.ReadTimeout),
		WriteTimeout:      getDurationEnv("WRITE_TIMEOUT", models.WriteTimeout),
		IdleTimeout:       getDurationEnv("IDLE_TIMEOUT", models.IdleTimeout),
		HeartbeatInterval: getDurationEnv("HEARTBEAT_INTERVAL", models.HeartbeatInterval),
		CleanupInterval:   getDurationEnv("CLEANUP_INTERVAL", models.CleanupInterval),
		ConnectionTimeout: getDurationEnv("CONNECTION_TIMEOUT", models.ConnectionTimeout),
		WebSocketTimeout:  getDurationEnv("WEBSOCKET_TIMEOUT", models.WebSocketTimeout),

		// Rate limiting settings
		MaxConnections:         getIntEnv(models.EnvMaxConnections, models.DefaultMaxConnections),
		HTTPRateLimitPerMinute: getIntEnv("HTTP_RATE_LIMIT_PER_MINUTE", models.DefaultHTTPRatePerMinute),
		WSRateLimitPerMinute:   getIntEnv("WS_RATE_LIMIT_PER_MINUTE", models.DefaultWSRatePerMinute),
		MaxWSConnPerIP:         getIntEnv("MAX_WS_CONN_PER_IP", models.DefaultMaxWSConnPerIP),

		// WebRTC settings
		STUNServers: getSTUNServers(),
		TURNServers: getTURNServers(),
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	return config
}

// getJWTSecret retrieves JWT secret from environment or generates a secure one
func getJWTSecret() []byte {
	if secret := os.Getenv(models.EnvJWTSecret); secret != "" {
		if len(secret) < models.MinJWTSecretLength {
			log.Fatalf("JWT secret must be at least %d characters long", models.MinJWTSecretLength)
		}
		return []byte(secret)
	}

	// Generate a secure secret if none provided (development only)
	env := getEnv(models.EnvEnvironment, models.EnvironmentDevelopment)
	if env == models.EnvironmentProduction {
		log.Fatal("JWT_SECRET environment variable is required in production")
	}

	// Generate secure random secret
	secret := make([]byte, 64)
	if _, err := rand.Read(secret); err != nil {
		log.Fatalf("Failed to generate secure JWT secret: %v", err)
	}

	log.Println("Warning: Using auto-generated JWT secret. Set JWT_SECRET environment variable for production.")
	return secret
}

// getAllowedOrigins parses the allowed origins from environment
func getAllowedOrigins() []string {
	originsEnv := getEnv(models.EnvAllowedOrigins, "")

	// If no origins specified, use defaults based on environment
	if originsEnv == "" {
		env := getEnv(models.EnvEnvironment, models.EnvironmentDevelopment)
		if env == models.EnvironmentProduction {
			log.Fatal("ALLOWED_ORIGINS environment variable is required in production")
		}
		return models.DefaultAllowedOrigins
	}

	// Parse comma-separated origins
	origins := strings.Split(originsEnv, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}

	return origins
}

// getSTUNServers parses STUN servers from environment
func getSTUNServers() []string {
	stunServers := getEnv("STUN_SERVERS", "")
	if stunServers == "" {
		return []string{
			models.DefaultSTUNServer1,
			models.DefaultSTUNServer2,
		}
	}

	servers := strings.Split(stunServers, ",")
	for i, server := range servers {
		servers[i] = strings.TrimSpace(server)
	}

	return servers
}

// getTURNServers parses TURN servers from environment
func getTURNServers() []TURNServerConfig {
	turnServers := getEnv("TURN_SERVERS", "")
	if turnServers == "" {
		return nil
	}

	var servers []TURNServerConfig
	serverConfigs := strings.Split(turnServers, ";")

	for _, config := range serverConfigs {
		parts := strings.Split(config, ",")
		if len(parts) == 3 {
			servers = append(servers, TURNServerConfig{
				URL:        strings.TrimSpace(parts[0]),
				Username:   strings.TrimSpace(parts[1]),
				Credential: strings.TrimSpace(parts[2]),
			})
		}
	}

	return servers
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	// Validate JWT secret length
	if len(config.JWTSecret) < models.MinJWTSecretLength {
		return fmt.Errorf("JWT secret must be at least %d characters long", models.MinJWTSecretLength)
	}

	// Validate environment
	validEnvironments := []string{
		models.EnvironmentDevelopment,
		models.EnvironmentStaging,
		models.EnvironmentProduction,
	}
	isValidEnv := false
	for _, env := range validEnvironments {
		if config.Environment == env {
			isValidEnv = true
			break
		}
	}
	if !isValidEnv {
		return fmt.Errorf("invalid environment: %s", config.Environment)
	}

	// Validate log level
	validLogLevels := []string{
		models.LogLevelDebug,
		models.LogLevelInfo,
		models.LogLevelWarn,
		models.LogLevelError,
		models.LogLevelFatal,
	}
	isValidLogLevel := false
	for _, level := range validLogLevels {
		if config.LogLevel == level {
			isValidLogLevel = true
			break
		}
	}
	if !isValidLogLevel {
		return fmt.Errorf("invalid log level: %s", config.LogLevel)
	}

	// Validate origins in production
	if config.Environment == models.EnvironmentProduction {
		for _, origin := range config.AllowedOrigins {
			if origin == "*" {
				return fmt.Errorf("wildcard origins not allowed in production")
			}
		}
	}

	return nil
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == models.EnvironmentProduction
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == models.EnvironmentDevelopment
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
