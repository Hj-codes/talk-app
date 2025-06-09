package utils

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear environment variables to test defaults
	envVars := []string{
		"PORT", "JWT_SECRET", "READ_TIMEOUT", "WRITE_TIMEOUT",
		"IDLE_TIMEOUT", "HEARTBEAT_INTERVAL", "CLEANUP_INTERVAL",
		"CONNECTION_TIMEOUT", "ALLOWED_ORIGINS", "MAX_CONNECTIONS",
		"RATE_LIMIT_PER_MINUTE",
	}

	// Store original values
	originalValues := make(map[string]string)
	for _, env := range envVars {
		originalValues[env] = os.Getenv(env)
		os.Unsetenv(env)
	}

	// Restore original values after test
	defer func() {
		for env, value := range originalValues {
			if value != "" {
				os.Setenv(env, value)
			}
		}
	}()

	config := LoadConfig()

	assert.Equal(t, "8080", config.Port)
	assert.Equal(t, []byte("my-secret-key-change-in-production"), config.JWTSecret)
	assert.Equal(t, 15*time.Second, config.ReadTimeout)
	assert.Equal(t, 15*time.Second, config.WriteTimeout)
	assert.Equal(t, 60*time.Second, config.IdleTimeout)
	assert.Equal(t, 30*time.Second, config.HeartbeatInterval)
	assert.Equal(t, 30*time.Second, config.CleanupInterval)
	assert.Equal(t, 60*time.Second, config.ConnectionTimeout)
	assert.Equal(t, []string{"*"}, config.AllowedOrigins)
	assert.Equal(t, 1000, config.MaxConnections)
	// assert.Equal(t, 60, config.RateLimitPerMinute)
}

func TestLoadConfig_EnvironmentVariables(t *testing.T) {
	// Set environment variables
	testEnvVars := map[string]string{
		"PORT":                  "9090",
		"JWT_SECRET":            "test-secret",
		"READ_TIMEOUT":          "20s",
		"WRITE_TIMEOUT":         "25s",
		"IDLE_TIMEOUT":          "120s",
		"HEARTBEAT_INTERVAL":    "45s",
		"CLEANUP_INTERVAL":      "60s",
		"CONNECTION_TIMEOUT":    "90s",
		"ALLOWED_ORIGINS":       "https://example.com",
		"MAX_CONNECTIONS":       "500",
		"RATE_LIMIT_PER_MINUTE": "30",
	}

	// Store original values and set test values
	originalValues := make(map[string]string)
	for env, value := range testEnvVars {
		originalValues[env] = os.Getenv(env)
		os.Setenv(env, value)
	}

	// Restore original values after test
	defer func() {
		for env, value := range originalValues {
			if value != "" {
				os.Setenv(env, value)
			} else {
				os.Unsetenv(env)
			}
		}
	}()

	config := LoadConfig()

	assert.Equal(t, "9090", config.Port)
	assert.Equal(t, []byte("test-secret"), config.JWTSecret)
	assert.Equal(t, 20*time.Second, config.ReadTimeout)
	assert.Equal(t, 25*time.Second, config.WriteTimeout)
	assert.Equal(t, 120*time.Second, config.IdleTimeout)
	assert.Equal(t, 45*time.Second, config.HeartbeatInterval)
	assert.Equal(t, 60*time.Second, config.CleanupInterval)
	assert.Equal(t, 90*time.Second, config.ConnectionTimeout)
	assert.Equal(t, []string{"https://example.com"}, config.AllowedOrigins)
	assert.Equal(t, 500, config.MaxConnections)
	// assert.Equal(t, 30, config.RateLimitPerMinute)
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "environment variable exists",
			key:          "TEST_ENV_VAR",
			defaultValue: "default",
			envValue:     "env_value",
			expected:     "env_value",
		},
		{
			name:         "environment variable does not exist",
			key:          "NON_EXISTENT_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "empty environment variable",
			key:          "EMPTY_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store original value
			original := os.Getenv(tt.key)
			defer func() {
				if original != "" {
					os.Setenv(tt.key, original)
				} else {
					os.Unsetenv(tt.key)
				}
			}()

			// Set test environment variable
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetIntEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		expected     int
	}{
		{
			name:         "valid integer",
			key:          "TEST_INT_VAR",
			defaultValue: 100,
			envValue:     "200",
			expected:     200,
		},
		{
			name:         "invalid integer",
			key:          "TEST_INVALID_INT",
			defaultValue: 100,
			envValue:     "not_a_number",
			expected:     100,
		},
		{
			name:         "empty environment variable",
			key:          "TEST_EMPTY_INT",
			defaultValue: 100,
			envValue:     "",
			expected:     100,
		},
		{
			name:         "zero value",
			key:          "TEST_ZERO_INT",
			defaultValue: 100,
			envValue:     "0",
			expected:     0,
		},
		{
			name:         "negative value",
			key:          "TEST_NEG_INT",
			defaultValue: 100,
			envValue:     "-50",
			expected:     -50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store original value
			original := os.Getenv(tt.key)
			defer func() {
				if original != "" {
					os.Setenv(tt.key, original)
				} else {
					os.Unsetenv(tt.key)
				}
			}()

			// Set test environment variable
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getIntEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDurationEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue time.Duration
		envValue     string
		expected     time.Duration
	}{
		{
			name:         "valid duration",
			key:          "TEST_DURATION_VAR",
			defaultValue: 30 * time.Second,
			envValue:     "45s",
			expected:     45 * time.Second,
		},
		{
			name:         "invalid duration",
			key:          "TEST_INVALID_DURATION",
			defaultValue: 30 * time.Second,
			envValue:     "not_a_duration",
			expected:     30 * time.Second,
		},
		{
			name:         "empty environment variable",
			key:          "TEST_EMPTY_DURATION",
			defaultValue: 30 * time.Second,
			envValue:     "",
			expected:     30 * time.Second,
		},
		{
			name:         "complex duration",
			key:          "TEST_COMPLEX_DURATION",
			defaultValue: 30 * time.Second,
			envValue:     "1h30m45s",
			expected:     1*time.Hour + 30*time.Minute + 45*time.Second,
		},
		{
			name:         "milliseconds",
			key:          "TEST_MS_DURATION",
			defaultValue: 30 * time.Second,
			envValue:     "500ms",
			expected:     500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store original value
			original := os.Getenv(tt.key)
			defer func() {
				if original != "" {
					os.Setenv(tt.key, original)
				} else {
					os.Unsetenv(tt.key)
				}
			}()

			// Set test environment variable
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getDurationEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test concurrent access to config loading
func TestLoadConfig_Concurrent(t *testing.T) {
	// This test ensures that concurrent calls to LoadConfig don't cause race conditions
	const numGoroutines = 10
	results := make(chan *Config, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			config := LoadConfig()
			results <- config
		}()
	}

	// Collect all results
	configs := make([]*Config, 0, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		configs = append(configs, <-results)
	}

	// All configs should have the same values
	firstConfig := configs[0]
	for i, config := range configs {
		assert.Equal(t, firstConfig.Port, config.Port, "Config %d port mismatch", i)
		assert.Equal(t, firstConfig.MaxConnections, config.MaxConnections, "Config %d max connections mismatch", i)
		assert.Equal(t, firstConfig.ReadTimeout, config.ReadTimeout, "Config %d read timeout mismatch", i)
	}
}

// Benchmark config loading
func BenchmarkLoadConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LoadConfig()
	}
}
