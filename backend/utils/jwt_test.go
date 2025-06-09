package utils

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateUUID(t *testing.T) {
	// Test that UUIDs are generated
	uuid1 := GenerateUUID()
	uuid2 := GenerateUUID()

	assert.NotEmpty(t, uuid1)
	assert.NotEmpty(t, uuid2)
	assert.NotEqual(t, uuid1, uuid2)

	// Test UUID format (should be 36 characters with hyphens)
	assert.Len(t, uuid1, 36)
	assert.Contains(t, uuid1, "-")

	// Test multiple generations don't collide
	uuids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		uuid := GenerateUUID()
		assert.False(t, uuids[uuid], "UUID collision detected: %s", uuid)
		uuids[uuid] = true
	}
}

func TestGenerateToken(t *testing.T) {
	userID := "test-user-123"

	token, err := GenerateToken(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Token should have 3 parts separated by dots (header.payload.signature)
	parts := strings.Split(token, ".")
	assert.Len(t, parts, 3)

	// Verify token can be parsed
	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	// Verify claims
	claims, ok := parsedToken.Claims.(*Claims)
	require.True(t, ok)
	assert.Equal(t, userID, claims.UserID)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
	assert.True(t, claims.IssuedAt.Before(time.Now().Add(time.Second)))
}

func TestValidateJWT_ValidToken(t *testing.T) {
	userID := "test-user-456"

	// Generate a token
	token, err := GenerateToken(userID)
	require.NoError(t, err)

	// Validate the token
	claims, err := ValidateJWT(token)
	require.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
}

func TestValidateJWT_InvalidToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "malformed token",
			token: "invalid.token.format",
		},
		{
			name:  "wrong signature",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdCIsImV4cCI6OTk5OTk5OTk5OX0.wrong_signature",
		},
		{
			name:  "not a JWT",
			token: "this-is-not-a-jwt-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateJWT(tt.token)
			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	userID := "test-user-expired"

	// Create an expired token manually
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	require.NoError(t, err)

	// Validate the expired token
	validatedClaims, err := ValidateJWT(tokenString)
	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
}

func TestJWT_RoundTrip(t *testing.T) {
	// Test multiple users
	userIDs := []string{
		"user1",
		"user-with-special-chars-123",
		"very-long-user-id-with-many-characters-to-test-edge-cases",
		"",
	}

	for _, userID := range userIDs {
		t.Run("user_"+userID, func(t *testing.T) {
			// Generate token
			token, err := GenerateToken(userID)
			require.NoError(t, err)

			// Validate token
			claims, err := ValidateJWT(token)
			require.NoError(t, err)
			assert.Equal(t, userID, claims.UserID)
		})
	}
}

func TestJWT_ConcurrentGeneration(t *testing.T) {
	const numGoroutines = 100
	const tokensPerGoroutine = 10

	var wg sync.WaitGroup
	tokens := make(chan string, numGoroutines*tokensPerGoroutine)
	errors := make(chan error, numGoroutines*tokensPerGoroutine)

	// Generate tokens concurrently
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < tokensPerGoroutine; j++ {
				userID := fmt.Sprintf("user-%d-%d", goroutineID, j)
				token, err := GenerateToken(userID)
				if err != nil {
					errors <- err
				} else {
					tokens <- token
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete, then close channels
	wg.Wait()
	close(tokens)
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Unexpected error during concurrent token generation: %v", err)
	}

	// Verify all tokens are unique and valid
	tokenSet := make(map[string]bool)
	tokenCount := 0
	for token := range tokens {
		assert.False(t, tokenSet[token], "Duplicate token generated: %s", token)
		tokenSet[token] = true
		tokenCount++

		// Validate each token
		claims, err := ValidateJWT(token)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
	}

	assert.Equal(t, numGoroutines*tokensPerGoroutine, tokenCount)
}

func TestClaims_Serialization(t *testing.T) {
	userID := "test-user"

	// Generate token
	token, err := GenerateToken(userID)
	require.NoError(t, err)

	// Parse token to get claims
	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	require.NoError(t, err)

	claims, ok := parsedToken.Claims.(*Claims)
	require.True(t, ok)

	// Test that claims have expected structure
	assert.Equal(t, userID, claims.UserID)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.True(t, claims.ExpiresAt.After(claims.IssuedAt.Time))
}

// Benchmark token generation
func BenchmarkGenerateToken(b *testing.B) {
	userID := "benchmark-user"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateToken(userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark token validation
func BenchmarkValidateJWT(b *testing.B) {
	userID := "benchmark-user"
	token, err := GenerateToken(userID)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ValidateJWT(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark UUID generation
func BenchmarkGenerateUUID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateUUID()
	}
}
