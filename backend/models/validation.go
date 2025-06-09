package models

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validation patterns
var (
	uuidPattern         = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
	sdpOfferPattern     = regexp.MustCompile(`^v=0\r?\n.*m=audio`)
	sdpAnswerPattern    = regexp.MustCompile(`^v=0\r?\n.*m=audio`)
	iceCandidatePattern = regexp.MustCompile(`^candidate:[a-zA-Z0-9+/]+`)
)

// Validator instance
var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom validators
	validate.RegisterValidation("uuid4", validateUUID)
	validate.RegisterValidation("sdp_offer", validateSDPOffer)
	validate.RegisterValidation("sdp_answer", validateSDPAnswer)
	validate.RegisterValidation("ice_candidate", validateICECandidate)
}

// ValidatedMessage represents a validated WebSocket message
type ValidatedMessage struct {
	Type    string      `json:"type" validate:"required,oneof=find_match offer answer ice_candidate call_start call_accept call_reject call_end ping pong disconnect get_ice_servers"`
	Payload interface{} `json:"payload" validate:"required"`
	From    string      `json:"from,omitempty" validate:"omitempty,uuid4"`
	To      string      `json:"to,omitempty" validate:"omitempty,uuid4"`
}

// ValidatedSDPMessage represents a validated SDP offer/answer
type ValidatedSDPMessage struct {
	Type string `json:"type" validate:"required,oneof=offer answer"`
	SDP  string `json:"sdp" validate:"required,min=10,max=10000"`
}

// ValidatedICECandidateMessage represents a validated ICE candidate
type ValidatedICECandidateMessage struct {
	Candidate     string `json:"candidate" validate:"required,ice_candidate,max=1000"`
	SDPMLineIndex int    `json:"sdpMLineIndex" validate:"min=0,max=10"`
	SDPMid        string `json:"sdpMid" validate:"required,max=100"`
}

// ValidatedCallMessage represents a validated call control message
type ValidatedCallMessage struct {
	Action string `json:"action" validate:"required,oneof=start accept reject end"`
	RoomID string `json:"room_id,omitempty" validate:"omitempty,uuid4"`
}

// ValidatedMediaInfo represents validated media information
type ValidatedMediaInfo struct {
	HasAudio bool   `json:"has_audio"`
	HasVideo bool   `json:"has_video"`
	Codec    string `json:"codec,omitempty" validate:"omitempty,oneof=opus g722 pcmu pcma vp8 vp9 h264"`
}

// ValidationError represents a validation error with details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value"`
}

// ValidateMessage validates a WebSocket message
func ValidateMessage(msg interface{}) error {
	return validate.Struct(msg)
}

// ValidateAndParseMessage validates and parses a raw message
func ValidateAndParseMessage(rawMsg map[string]interface{}) (*ValidatedMessage, error) {
	// Extract and validate message type
	msgType, ok := rawMsg["type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid message type")
	}

	// Create base message
	validMsg := &ValidatedMessage{
		Type:    msgType,
		Payload: rawMsg["payload"],
	}

	// Extract optional fields
	if from, ok := rawMsg["from"].(string); ok {
		validMsg.From = from
	}
	if to, ok := rawMsg["to"].(string); ok {
		validMsg.To = to
	}

	// Validate the message structure
	if err := validate.Struct(validMsg); err != nil {
		return nil, formatValidationError(err)
	}

	// Type-specific payload validation
	switch msgType {
	case "offer", "answer":
		if err := validateSDPPayload(rawMsg["payload"]); err != nil {
			return nil, err
		}
	case "ice_candidate":
		if err := validateICEPayload(rawMsg["payload"]); err != nil {
			return nil, err
		}
	case "call_start", "call_accept", "call_reject", "call_end":
		if err := validateCallPayload(rawMsg["payload"]); err != nil {
			return nil, err
		}
	}

	return validMsg, nil
}

// validateSDPPayload validates SDP offer/answer payload
func validateSDPPayload(payload interface{}) error {
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid SDP payload format")
	}

	sdpMsg := &ValidatedSDPMessage{}
	if msgType, ok := payloadMap["type"].(string); ok {
		sdpMsg.Type = msgType
	}
	if sdp, ok := payloadMap["sdp"].(string); ok {
		sdpMsg.SDP = sdp
	}

	if err := validate.Struct(sdpMsg); err != nil {
		return formatValidationError(err)
	}

	// Additional SDP content validation
	if sdpMsg.Type == "offer" && !sdpOfferPattern.MatchString(sdpMsg.SDP) {
		return fmt.Errorf("invalid SDP offer format")
	}
	if sdpMsg.Type == "answer" && !sdpAnswerPattern.MatchString(sdpMsg.SDP) {
		return fmt.Errorf("invalid SDP answer format")
	}

	return nil
}

// validateICEPayload validates ICE candidate payload
func validateICEPayload(payload interface{}) error {
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid ICE candidate payload format")
	}

	iceMsg := &ValidatedICECandidateMessage{}
	if candidate, ok := payloadMap["candidate"].(string); ok {
		iceMsg.Candidate = candidate
	}
	if sdpMLineIndex, ok := payloadMap["sdpMLineIndex"].(float64); ok {
		iceMsg.SDPMLineIndex = int(sdpMLineIndex)
	}
	if sdpMid, ok := payloadMap["sdpMid"].(string); ok {
		iceMsg.SDPMid = sdpMid
	}

	return validate.Struct(iceMsg)
}

// validateCallPayload validates call control payload
func validateCallPayload(payload interface{}) error {
	if payload == nil {
		return nil // Call messages can have empty payload
	}

	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid call payload format")
	}

	callMsg := &ValidatedCallMessage{}
	if action, ok := payloadMap["action"].(string); ok {
		callMsg.Action = action
	}
	if roomID, ok := payloadMap["room_id"].(string); ok {
		callMsg.RoomID = roomID
	}

	return validate.Struct(callMsg)
}

// Custom validators
func validateUUID(fl validator.FieldLevel) bool {
	return uuidPattern.MatchString(fl.Field().String())
}

func validateSDPOffer(fl validator.FieldLevel) bool {
	sdp := fl.Field().String()
	return sdpOfferPattern.MatchString(sdp)
}

func validateSDPAnswer(fl validator.FieldLevel) bool {
	sdp := fl.Field().String()
	return sdpAnswerPattern.MatchString(sdp)
}

func validateICECandidate(fl validator.FieldLevel) bool {
	candidate := fl.Field().String()
	return iceCandidatePattern.MatchString(candidate)
}

// formatValidationError formats validation errors for better readability
func formatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errors []ValidationError
		for _, validationError := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   validationError.Field(),
				Message: getValidationMessage(validationError),
				Value:   fmt.Sprintf("%v", validationError.Value()),
			})
		}
		return fmt.Errorf("validation failed: %v", errors)
	}
	return err
}

// getValidationMessage returns a human-readable validation message
func getValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "uuid4":
		return "Must be a valid UUID"
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", err.Param())
	case "min":
		return fmt.Sprintf("Must be at least %s characters", err.Param())
	case "max":
		return fmt.Sprintf("Must be at most %s characters", err.Param())
	case "sdp_offer":
		return "Must be a valid SDP offer"
	case "sdp_answer":
		return "Must be a valid SDP answer"
	case "ice_candidate":
		return "Must be a valid ICE candidate"
	default:
		return fmt.Sprintf("Validation failed: %s", err.Tag())
	}
}

// SanitizeString removes potentially dangerous characters from strings
func SanitizeString(input string) string {
	// Remove null bytes, control characters, and excessive whitespace
	cleaned := strings.ReplaceAll(input, "\x00", "")
	cleaned = regexp.MustCompile(`[\x00-\x1f\x7f]`).ReplaceAllString(cleaned, "")
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}

// ValidateMessageSize checks if message size is within limits
func ValidateMessageSize(data []byte) error {
	const maxMessageSize = 64 * 1024 // 64KB
	if len(data) > maxMessageSize {
		return fmt.Errorf("message size exceeds maximum allowed size of %d bytes", maxMessageSize)
	}
	return nil
}
