package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"time"
	"voice-chat-app/models"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var (
	jwtSecret      []byte
	tokenBlacklist = make(map[string]time.Time)
	blacklistMutex sync.RWMutex
	configOnce     sync.Once
)

type Claims struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// Error definitions
var (
	ErrTokenBlacklisted = errors.New("token has been revoked")
	ErrTokenExpired     = errors.New("token has expired")
	ErrInvalidToken     = errors.New("invalid token")
	ErrWeakSecret       = errors.New("JWT secret is too weak")
)

// initJWTConfig initializes JWT configuration once
func initJWTConfig() {
	configOnce.Do(func() {
		config := LoadConfig()
		jwtSecret = config.JWTSecret

		// Validate JWT secret strength
		if len(jwtSecret) < models.MinJWTSecretLength {
			panic(ErrWeakSecret)
		}

		// Start blacklist cleanup goroutine
		go cleanupBlacklist()
	})
}

// ensureJWTInit ensures JWT configuration is initialized
func ensureJWTInit() {
	if jwtSecret == nil {
		initJWTConfig()
	}
}

func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateSecureSecret generates a cryptographically secure secret
func GenerateSecureSecret() ([]byte, error) {
	secret := make([]byte, 64) // 512 bits
	_, err := rand.Read(secret)
	return secret, err
}

func GenerateToken(userID string) (string, error) {
	ensureJWTInit()

	sessionID := GenerateUUID()

	claims := &Claims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "voice-chat-app",
			Subject:   userID,
			ID:        sessionID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenString string) (*Claims, error) {
	ensureJWTInit()

	// Check if token is blacklisted
	if isTokenBlacklisted(tokenString) {
		return nil, ErrTokenBlacklisted
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, ErrTokenExpired
			}
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Additional claims validation
	if claims.UserID == "" || claims.SessionID == "" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RevokeToken adds a token to the blacklist
func RevokeToken(tokenString string) error {
	ensureJWTInit()

	// Parse token to get expiration time
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return err
	}

	blacklistMutex.Lock()
	defer blacklistMutex.Unlock()

	// Add token to blacklist until its expiration
	if claims.ExpiresAt != nil {
		tokenBlacklist[tokenString] = claims.ExpiresAt.Time
	}

	return nil
}

// isTokenBlacklisted checks if a token is in the blacklist
func isTokenBlacklisted(tokenString string) bool {
	blacklistMutex.RLock()
	defer blacklistMutex.RUnlock()

	expiry, exists := tokenBlacklist[tokenString]
	if !exists {
		return false
	}

	// Check if token is still within its expiration time
	return time.Now().Before(expiry)
}

// cleanupBlacklist removes expired tokens from the blacklist
func cleanupBlacklist() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		blacklistMutex.Lock()
		now := time.Now()

		for token, expiry := range tokenBlacklist {
			if now.After(expiry) {
				delete(tokenBlacklist, token)
			}
		}

		blacklistMutex.Unlock()
	}
}

// RefreshToken generates a new token for a user if the current token is valid
func RefreshToken(tokenString string) (string, error) {
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return "", err
	}

	// Check if token is close to expiry (within 1 hour)
	if claims.ExpiresAt != nil && time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", errors.New("token does not need refresh yet")
	}

	// Revoke old token
	RevokeToken(tokenString)

	// Generate new token
	return GenerateToken(claims.UserID)
}

// GetBlacklistStats returns statistics about the token blacklist
func GetBlacklistStats() map[string]interface{} {
	blacklistMutex.RLock()
	defer blacklistMutex.RUnlock()

	return map[string]interface{}{
		"blacklisted_tokens": len(tokenBlacklist),
	}
}
