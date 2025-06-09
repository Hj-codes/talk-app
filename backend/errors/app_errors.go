package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"voice-chat-app/models"
)

// AppError represents a structured application error
type AppError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	StatusCode int                    `json:"-"`
	Cause      error                  `json:"-"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for error unwrapping
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithContext adds context information to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithCause sets the underlying cause of the error
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// ToJSON converts the error to JSON format
func (e *AppError) ToJSON() []byte {
	data, _ := json.Marshal(e)
	return data
}

// Pre-defined error constructors

// NewValidationError creates a new validation error
func NewValidationError(message string, details ...string) *AppError {
	err := &AppError{
		Code:       models.ErrorCodeValidation,
		Message:    message,
		StatusCode: models.StatusValidationFailed,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:       models.ErrorCodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
	}
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string) *AppError {
	if message == "" {
		message = "Unauthorized access"
	}
	return &AppError{
		Code:       models.ErrorCodeUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(message string) *AppError {
	if message == "" {
		message = "Rate limit exceeded"
	}
	return &AppError{
		Code:       models.ErrorCodeRateLimit,
		Message:    message,
		StatusCode: models.StatusRateLimited,
	}
}

// NewInternalError creates a new internal server error
func NewInternalError(message string, cause error) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return &AppError{
		Code:       models.ErrorCodeInternalError,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Cause:      cause,
	}
}

// NewInvalidMessageError creates a new invalid message error
func NewInvalidMessageError(messageType string) *AppError {
	return &AppError{
		Code:       models.ErrorCodeInvalidMessage,
		Message:    fmt.Sprintf("Invalid message type: %s", messageType),
		StatusCode: http.StatusBadRequest,
	}
}

// NewNoPartnerError creates a new no partner available error
func NewNoPartnerError() *AppError {
	return &AppError{
		Code:       models.ErrorCodeNoPartner,
		Message:    "No partner available for matching",
		StatusCode: http.StatusServiceUnavailable,
	}
}

// NewConnectionLostError creates a new connection lost error
func NewConnectionLostError(userID string) *AppError {
	err := &AppError{
		Code:       models.ErrorCodeConnectionLost,
		Message:    "Connection lost",
		StatusCode: http.StatusGone,
	}
	return err.WithContext("user_id", userID)
}

// NewInvalidStateError creates a new invalid state error
func NewInvalidStateError(currentState, expectedState string) *AppError {
	return &AppError{
		Code:       models.ErrorCodeInvalidState,
		Message:    fmt.Sprintf("Invalid state transition from %s to %s", currentState, expectedState),
		StatusCode: http.StatusConflict,
		Context: map[string]interface{}{
			"current_state":  currentState,
			"expected_state": expectedState,
		},
	}
}

// Error handling middleware

// ErrorHandler is a middleware that handles application errors
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				var appErr *AppError

				switch e := err.(type) {
				case *AppError:
					appErr = e
				case error:
					appErr = NewInternalError("Panic occurred", e)
				default:
					appErr = NewInternalError("Unknown panic occurred", fmt.Errorf("%v", e))
				}

				WriteErrorResponse(w, appErr)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// WriteErrorResponse writes an error response to the HTTP response writer
func WriteErrorResponse(w http.ResponseWriter, err *AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)
	w.Write(err.ToJSON())
}

// WebSocket error handling

// WebSocketErrorResponse represents a WebSocket error response
type WebSocketErrorResponse struct {
	Type      string                 `json:"type"`
	Error     string                 `json:"error"`
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   string                 `json:"details,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

// ToWebSocketError converts an AppError to a WebSocket error format
func (e *AppError) ToWebSocketError() *WebSocketErrorResponse {
	return &WebSocketErrorResponse{
		Type:      models.MessageTypeError,
		Error:     "application_error",
		Code:      e.Code,
		Message:   e.Message,
		Details:   e.Details,
		Context:   e.Context,
		Timestamp: fmt.Sprintf("%d", time.Now().Unix()),
	}
}

// Error utilities

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == models.ErrorCodeValidation
	}
	return false
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == models.ErrorCodeNotFound
	}
	return false
}

// IsRateLimitError checks if an error is a rate limit error
func IsRateLimitError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == models.ErrorCodeRateLimit
	}
	return false
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return models.ErrorCodeInternalError
}

// GetStatusCode extracts the HTTP status code from an error
func GetStatusCode(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}
