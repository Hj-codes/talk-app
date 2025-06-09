package models

import "time"

// User status constants
const (
	StatusWaiting      = "waiting"
	StatusConnected    = "connected"
	StatusDisconnected = "disconnected"
	StatusMatched      = "matched"
)

// WebSocket message types
const (
	MessageTypeFindMatch     = "find_match"
	MessageTypeOffer         = "offer"
	MessageTypeAnswer        = "answer"
	MessageTypeICECandidate  = "ice_candidate"
	MessageTypeCallStart     = "call_start"
	MessageTypeCallAccept    = "call_accept"
	MessageTypeCallReject    = "call_reject"
	MessageTypeCallEnd       = "call_end"
	MessageTypePing          = "ping"
	MessageTypePong          = "pong"
	MessageTypeDisconnect    = "disconnect"
	MessageTypeGetICEServers = "get_ice_servers"
	MessageTypeSession       = "session"
	MessageTypeUserMatched   = "user_matched"
	MessageTypeUserLeft      = "user_left"
	MessageTypeError         = "error"
)

// Call states
const (
	CallStateIdle     = "idle"
	CallStateRinging  = "ringing"
	CallStateAnswered = "answered"
	CallStateEnded    = "ended"
	CallStateFailed   = "failed"
)

// Error codes
const (
	ErrorCodeValidation     = "VALIDATION_ERROR"
	ErrorCodeNotFound       = "NOT_FOUND"
	ErrorCodeUnauthorized   = "UNAUTHORIZED"
	ErrorCodeRateLimit      = "RATE_LIMIT_EXCEEDED"
	ErrorCodeInternalError  = "INTERNAL_ERROR"
	ErrorCodeInvalidMessage = "INVALID_MESSAGE"
	ErrorCodeNoPartner      = "NO_PARTNER"
	ErrorCodeConnectionLost = "CONNECTION_LOST"
	ErrorCodeInvalidState   = "INVALID_STATE"
)

// Timeout constants
const (
	ConnectionTimeout   = 60 * time.Second
	HeartbeatInterval   = 30 * time.Second
	CleanupInterval     = 30 * time.Second
	WebSocketTimeout    = 45 * time.Second
	ReadTimeout         = 15 * time.Second
	WriteTimeout        = 15 * time.Second
	IdleTimeout         = 60 * time.Second
	TokenExpiryDuration = 24 * time.Hour
	TokenRefreshWindow  = 1 * time.Hour
)

// Size limits
const (
	MaxMessageSize      = 64 * 1024 // 64KB
	MaxSDPSize          = 10 * 1024 // 10KB
	MaxICECandidateSize = 1024      // 1KB
	MaxUserIDLength     = 100
	MaxSessionIDLength  = 100
	MaxRoomIDLength     = 100
)

// Rate limiting constants
const (
	DefaultHTTPRatePerMinute = 60
	DefaultWSRatePerMinute   = 100
	DefaultMaxWSConnPerIP    = 10
	DefaultMaxConnections    = 1000
)

// WebRTC constants
const (
	DefaultSTUNServer1 = "stun:stun.l.google.com:19302"
	DefaultSTUNServer2 = "stun:stun1.l.google.com:19302"
)

// Default allowed origins for development
var DefaultAllowedOrigins = []string{
	"http://localhost:3000",
	"http://localhost:8080",
	"https://localhost:3000",
	"https://localhost:8080",
}

// Production security settings
const (
	MinJWTSecretLength = 8
	MinPasswordLength  = 8
	MaxRetryAttempts   = 3
	LockoutDuration    = 15 * time.Minute
)

// Media codec constants
const (
	CodecOpus = "opus"
	CodecG722 = "g722"
	CodecPCMU = "pcmu"
	CodecPCMA = "pcma"
	CodecVP8  = "vp8"
	CodecVP9  = "vp9"
	CodecH264 = "h264"
)

// HTTP status constants for custom responses
const (
	StatusValidationFailed = 422
	StatusRateLimited      = 429
	StatusUpgradeRequired  = 426
)

// Validation patterns
const (
	UUIDPattern         = `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`
	SDPOfferPattern     = `^v=0\r?\n.*m=audio`
	SDPAnswerPattern    = `^v=0\r?\n.*m=audio`
	ICECandidatePattern = `^candidate:[a-zA-Z0-9+/]+`
)

// Environment variable names
const (
	EnvPort               = "PORT"
	EnvJWTSecret          = "JWT_SECRET"
	EnvAllowedOrigins     = "ALLOWED_ORIGINS"
	EnvMaxConnections     = "MAX_CONNECTIONS"
	EnvRateLimitPerMinute = "RATE_LIMIT_PER_MINUTE"
	EnvLogLevel           = "LOG_LEVEL"
	EnvEnvironment        = "ENVIRONMENT"
)

// Log levels
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)

// Environment types
const (
	EnvironmentDevelopment = "development"
	EnvironmentStaging     = "staging"
	EnvironmentProduction  = "production"
)
