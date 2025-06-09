package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"voice-chat-app/models"
	"voice-chat-app/utils"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, implement proper CORS
	},
	HandshakeTimeout: 45 * time.Second,
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
}

type SignalingServer struct {
	UserPool    *models.UserPool
	RateLimiter interface{} // Will be updated to proper type later
	STUNServers []string
	TURNServers []TURNServer
}

type TURNServer struct {
	URL        string `json:"url"`
	Username   string `json:"username"`
	Credential string `json:"credential"`
}

type ICEServersResponse struct {
	ICEServers []ICEServer `json:"iceServers"`
}

type ICEServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

type Message struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	From      string      `json:"from,omitempty"`
	To        string      `json:"to,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type WebRTCMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// SDP validation regex patterns
var (
	sdpOfferPattern  = regexp.MustCompile(`^v=0\r?\n.*m=audio`)
	sdpAnswerPattern = regexp.MustCompile(`^v=0\r?\n.*m=audio`)
)

func (s *SignalingServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DEBUG] WebSocket upgrade attempt from %s", r.RemoteAddr)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	log.Printf("[DEBUG] WebSocket connection established from %s", r.RemoteAddr)

	// Generate user session
	userID := utils.GenerateUUID()
	token, err := utils.GenerateToken(userID)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		conn.Close()
		return
	}

	log.Printf("[DEBUG] Generated session for user %s with token", userID)

	// Create connection wrapper
	connection := &models.Connection{
		Conn:     conn,
		UserID:   userID,
		LastPing: time.Now(),
		IsActive: true,
	}

	user := &models.User{
		ID:         userID,
		SessionID:  token,
		Status:     "waiting",
		Connection: connection,
	}

	// Send session info to client
	sessionMsg := Message{
		Type:      "session",
		Timestamp: time.Now(),
		Payload: map[string]string{
			"user_id": userID,
			"token":   token,
		},
	}

	if err := connection.WriteJSON(sessionMsg); err != nil {
		log.Printf("Error sending session message: %v", err)
		connection.Close()
		return
	}

	log.Printf("[DEBUG] Session message sent to user %s", userID)

	s.UserPool.AddWaitingUser(user)
	log.Printf("[DEBUG] User %s added to waiting pool", userID)

	// Get current stats
	stats := s.UserPool.GetStats()
	log.Printf("[DEBUG] Current server stats - Waiting: %d, Active: %d, Rooms: %d",
		stats["waiting_users"], stats["active_users"], stats["active_rooms"])

	// Start heartbeat goroutine
	go s.handleHeartbeat(connection)

	// Handle user messages
	s.handleUserMessages(connection, user)

	// Cleanup on disconnect
	log.Printf("[DEBUG] User %s connection ended, cleaning up", userID)
	s.handleDisconnect(user)
}

func (s *SignalingServer) handleHeartbeat(conn *models.Connection) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !conn.IsActive {
				return
			}

			pingMsg := Message{
				Type:      "ping",
				Timestamp: time.Now(),
			}

			if err := conn.WriteJSON(pingMsg); err != nil {
				log.Printf("Heartbeat failed for user %s: %v", conn.UserID, err)
				conn.Close()
				return
			}
		}
	}
}

func (s *SignalingServer) handleUserMessages(conn *models.Connection, user *models.User) {
	// Set read deadline
	conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		var msg Message
		err := conn.Conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Read error for user %s: %v", user.ID, err)
			break
		}

		// Update ping time and reset read deadline
		conn.UpdatePing()
		conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		// Add timestamp and from field
		msg.Timestamp = time.Now()
		msg.From = user.ID

		// Log all incoming messages for debugging
		log.Printf("[DEBUG] Received message from user %s: type=%s", user.ID, msg.Type)

		switch msg.Type {
		case "pong":
			// Handle pong response - just update ping time (already done above)
			log.Printf("[DEBUG] Pong received from user %s", user.ID)
			continue
		case "find_match":
			log.Printf("[DEBUG] User %s requesting match", user.ID)
			s.handleFindMatch(user)
		case "offer":
			log.Printf("[DEBUG] WebRTC offer received from user %s", user.ID)
			s.handleWebRTCOffer(msg, user)
		case "answer":
			log.Printf("[DEBUG] WebRTC answer received from user %s", user.ID)
			s.handleWebRTCAnswer(msg, user)
		case "ice_candidate":
			log.Printf("[DEBUG] ICE candidate received from user %s", user.ID)
			s.handleICECandidate(msg, user)
		case "call_start":
			log.Printf("[DEBUG] Call start request from user %s", user.ID)
			s.handleCallStart(msg, user)
		case "call_end":
			log.Printf("[DEBUG] Call end request from user %s", user.ID)
			s.handleCallEnd(msg, user)
		case "call_accept":
			log.Printf("[DEBUG] Call accept from user %s", user.ID)
			s.handleCallAccept(msg, user)
		case "call_reject":
			log.Printf("[DEBUG] Call reject from user %s", user.ID)
			s.handleCallReject(msg, user)
		case "get_ice_servers":
			log.Printf("[DEBUG] ICE servers request from user %s", user.ID)
			s.handleGetICEServers(user)
		case "disconnect":
			log.Printf("[DEBUG] User %s disconnecting", user.ID)
			return // Exit the loop to trigger cleanup
		default:
			log.Printf("Unknown message type: %s from user %s", msg.Type, user.ID)
		}
	}
}

func (s *SignalingServer) relaySignaling(msg Message) {
	// Find the target user and relay the signaling message
	if msg.To == "" {
		log.Printf("No target specified for signaling message from %s", msg.From)
		return
	}

	targetUser := s.UserPool.GetActiveUser(msg.To)
	if targetUser == nil {
		log.Printf("Target user %s not found for message from %s", msg.To, msg.From)
		return
	}

	// Verify users are in the same room
	senderUser := s.UserPool.GetActiveUser(msg.From)
	if senderUser == nil || senderUser.RoomID != targetUser.RoomID || senderUser.RoomID == "" {
		log.Printf("Users %s and %s are not in the same room", msg.From, msg.To)
		return
	}

	// Relay the message to the target user
	if err := targetUser.Connection.WriteJSON(msg); err != nil {
		log.Printf("Error relaying message to user %s: %v", msg.To, err)
	}
}

func (s *SignalingServer) handleFindMatch(user *models.User) {
	log.Printf("[DEBUG] Processing find match request for user %s", user.ID)

	partner := s.UserPool.GetRandomWaitingUser(user.ID)
	if partner == nil {
		log.Printf("[DEBUG] No partner found for user %s, sending waiting status", user.ID)
		// No match found, send waiting status
		waitingMsg := Message{
			Type:      "waiting",
			Timestamp: time.Now(),
			Payload: map[string]string{
				"status": "Looking for a partner...",
			},
		}
		if err := user.Connection.WriteJSON(waitingMsg); err != nil {
			log.Printf("[ERROR] Failed to send waiting message to user %s: %v", user.ID, err)
		}
		return
	}

	log.Printf("[DEBUG] Found partner %s for user %s, creating room", partner.ID, user.ID)

	// Create room for both users
	room := s.UserPool.CreateRoom(user, partner)

	log.Printf("[DEBUG] Created room %s for users %s (caller) and %s (callee)", room.ID, user.ID, partner.ID)

	// Notify both users of the match
	matchMsg := Message{
		Type:      "match_found",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"partner_id": partner.ID,
			"room_id":    room.ID,
			"role":       "caller", // User who initiated gets caller role
		},
	}

	partnerMatchMsg := Message{
		Type:      "match_found",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"partner_id": user.ID,
			"room_id":    room.ID,
			"role":       "callee", // Partner gets callee role
		},
	}

	if err := user.Connection.WriteJSON(matchMsg); err != nil {
		log.Printf("Error notifying user %s of match: %v", user.ID, err)
		return
	}

	if err := partner.Connection.WriteJSON(partnerMatchMsg); err != nil {
		log.Printf("Error notifying partner %s of match: %v", partner.ID, err)
		return
	}

	log.Printf("Successfully created room %s and notified both users: %s (caller) and %s (callee)", room.ID, user.ID, partner.ID)
}

func (s *SignalingServer) handleDisconnect(user *models.User) {
	log.Printf("[DEBUG] Starting disconnect process for user %s", user.ID)

	// Find partner and notify them
	partner := s.UserPool.FindPartner(user.ID)
	if partner != nil {
		log.Printf("[DEBUG] Found partner %s for disconnecting user %s, notifying", partner.ID, user.ID)

		disconnectMsg := Message{
			Type:      "partner_disconnected",
			Timestamp: time.Now(),
			Payload: map[string]string{
				"reason": "Partner left the conversation",
			},
		}

		if err := partner.Connection.WriteJSON(disconnectMsg); err != nil {
			log.Printf("Error notifying partner of disconnect: %v", err)
		} else {
			log.Printf("[DEBUG] Partner %s notified of user %s disconnection", partner.ID, user.ID)
		}

		// Move partner back to waiting
		s.UserPool.MoveToWaiting(partner.ID)
		log.Printf("[DEBUG] Moved partner %s back to waiting pool", partner.ID)
	} else {
		log.Printf("[DEBUG] No partner found for disconnecting user %s", user.ID)
	}

	// Remove user from pools and close connection
	s.UserPool.RemoveUser(user.ID)
	user.Connection.Close()

	// Get updated stats
	stats := s.UserPool.GetStats()
	log.Printf("[DEBUG] User %s cleanup complete. Updated stats - Waiting: %d, Active: %d, Rooms: %d",
		user.ID, stats["waiting_users"], stats["active_users"], stats["active_rooms"])
}

// WebRTC-specific handlers

func (s *SignalingServer) handleWebRTCOffer(msg Message, user *models.User) {
	log.Printf("[DEBUG] Processing WebRTC offer from user %s", user.ID)

	// Log the payload structure for debugging
	if payload, ok := msg.Payload.(map[string]interface{}); ok {
		log.Printf("[DEBUG] Offer payload structure: %+v", payload)
		if sdp, exists := payload["sdp"]; exists {
			if sdpStr, ok := sdp.(string); ok {
				log.Printf("[DEBUG] SDP offer length: %d characters", len(sdpStr))
				log.Printf("[DEBUG] SDP offer preview: %.200s...", sdpStr)
			}
		}
	} else {
		log.Printf("[ERROR] Offer payload is not a map: %T", msg.Payload)
	}

	validationResult, errorMsg := s.validateSDPOfferDetailed(msg.Payload)
	if !validationResult {
		log.Printf("[ERROR] SDP offer validation failed for user %s: %s", user.ID, errorMsg)
		s.sendError(user, fmt.Sprintf("Invalid SDP offer format: %s", errorMsg))
		return
	}

	log.Printf("[DEBUG] SDP offer validation passed for user %s", user.ID)

	partner := s.UserPool.FindPartner(user.ID)
	if partner == nil {
		log.Printf("[ERROR] No partner found for WebRTC offer from user %s", user.ID)
		s.sendError(user, "No partner found for WebRTC offer")
		return
	}

	log.Printf("[DEBUG] Found partner %s for offer from user %s", partner.ID, user.ID)

	// Update call state
	user.CallState = models.CallStateRinging
	partner.CallState = models.CallStateRinging

	// Forward offer to partner
	msg.To = partner.ID
	msg.From = user.ID
	if err := partner.Connection.WriteJSON(msg); err != nil {
		log.Printf("Error forwarding offer to partner %s: %v", partner.ID, err)
		s.sendError(user, "Failed to forward offer")
		return
	}

	log.Printf("WebRTC offer successfully forwarded from %s to %s", user.ID, partner.ID)
}

func (s *SignalingServer) handleWebRTCAnswer(msg Message, user *models.User) {
	log.Printf("[DEBUG] Processing WebRTC answer from user %s", user.ID)

	// Log the payload structure for debugging
	if payload, ok := msg.Payload.(map[string]interface{}); ok {
		log.Printf("[DEBUG] Answer payload structure: %+v", payload)
		if sdp, exists := payload["sdp"]; exists {
			if sdpStr, ok := sdp.(string); ok {
				log.Printf("[DEBUG] SDP answer length: %d characters", len(sdpStr))
				log.Printf("[DEBUG] SDP answer preview: %.200s...", sdpStr)
			}
		}
	} else {
		log.Printf("[ERROR] Answer payload is not a map: %T", msg.Payload)
	}

	validationResult, errorMsg := s.validateSDPAnswerDetailed(msg.Payload)
	if !validationResult {
		log.Printf("[ERROR] SDP answer validation failed for user %s: %s", user.ID, errorMsg)
		s.sendError(user, fmt.Sprintf("Invalid SDP answer format: %s", errorMsg))
		return
	}

	log.Printf("[DEBUG] SDP answer validation passed for user %s", user.ID)

	partner := s.UserPool.FindPartner(user.ID)
	if partner == nil {
		log.Printf("[ERROR] No partner found for WebRTC answer from user %s", user.ID)
		s.sendError(user, "No partner found for WebRTC answer")
		return
	}

	log.Printf("[DEBUG] Found partner %s for answer from user %s", partner.ID, user.ID)

	// Update call state
	user.CallState = models.CallStateAnswered
	partner.CallState = models.CallStateAnswered

	// Update room call state
	if roomID := user.RoomID; roomID != "" {
		if room := s.UserPool.Rooms[roomID]; room != nil {
			room.CallState = models.CallStateAnswered
			now := time.Now()
			room.StartedAt = &now
			log.Printf("[DEBUG] Updated room %s call state to answered", roomID)
		}
	}

	// Forward answer to partner
	msg.To = partner.ID
	msg.From = user.ID
	if err := partner.Connection.WriteJSON(msg); err != nil {
		log.Printf("Error forwarding answer to partner %s: %v", partner.ID, err)
		s.sendError(user, "Failed to forward answer")
		return
	}

	log.Printf("WebRTC answer successfully forwarded from %s to %s", user.ID, partner.ID)
}

func (s *SignalingServer) handleICECandidate(msg Message, user *models.User) {
	log.Printf("[DEBUG] Processing ICE candidate from user %s", user.ID)

	// Log the payload structure for debugging
	if payload, ok := msg.Payload.(map[string]interface{}); ok {
		log.Printf("[DEBUG] ICE candidate payload: %+v", payload)
	} else {
		log.Printf("[ERROR] ICE candidate payload is not a map: %T", msg.Payload)
	}

	if !s.validateICECandidate(msg.Payload) {
		log.Printf("[ERROR] ICE candidate validation failed for user %s", user.ID)
		s.sendError(user, "Invalid ICE candidate format")
		return
	}

	log.Printf("[DEBUG] ICE candidate validation passed for user %s", user.ID)

	partner := s.UserPool.FindPartner(user.ID)
	if partner == nil {
		log.Printf("No partner found for ICE candidate from user %s", user.ID)
		return
	}

	log.Printf("[DEBUG] Forwarding ICE candidate from user %s to partner %s", user.ID, partner.ID)

	// Forward ICE candidate to partner
	msg.To = partner.ID
	msg.From = user.ID
	if err := partner.Connection.WriteJSON(msg); err != nil {
		log.Printf("Error forwarding ICE candidate to partner %s: %v", partner.ID, err)
		return
	}

	log.Printf("[DEBUG] ICE candidate successfully forwarded from %s to %s", user.ID, partner.ID)
}

func (s *SignalingServer) handleCallStart(msg Message, user *models.User) {
	partner := s.UserPool.FindPartner(user.ID)
	if partner == nil {
		s.sendError(user, "No partner found to start call")
		return
	}

	// Send call_incoming to partner
	callMsg := Message{
		Type:      "call_incoming",
		From:      user.ID,
		To:        partner.ID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"caller_id": user.ID,
			"room_id":   user.RoomID,
		},
	}

	if err := partner.Connection.WriteJSON(callMsg); err != nil {
		log.Printf("Error sending call_incoming to partner %s: %v", partner.ID, err)
		s.sendError(user, "Failed to initiate call")
		return
	}

	user.CallState = models.CallStateRinging
	log.Printf("Call initiated from %s to %s", user.ID, partner.ID)
}

func (s *SignalingServer) handleCallAccept(msg Message, user *models.User) {
	partner := s.UserPool.FindPartner(user.ID)
	if partner == nil {
		s.sendError(user, "No partner found to accept call")
		return
	}

	// Send call_accepted to partner
	acceptMsg := Message{
		Type:      "call_accepted",
		From:      user.ID,
		To:        partner.ID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"callee_id": user.ID,
			"room_id":   user.RoomID,
		},
	}

	if err := partner.Connection.WriteJSON(acceptMsg); err != nil {
		log.Printf("Error sending call_accepted to partner %s: %v", partner.ID, err)
		return
	}

	user.CallState = models.CallStateAnswered
	partner.CallState = models.CallStateAnswered

	log.Printf("Call accepted by %s from %s", user.ID, partner.ID)
}

func (s *SignalingServer) handleCallReject(msg Message, user *models.User) {
	partner := s.UserPool.FindPartner(user.ID)
	if partner == nil {
		return
	}

	// Send call_rejected to partner
	rejectMsg := Message{
		Type:      "call_rejected",
		From:      user.ID,
		To:        partner.ID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"reason": "Call rejected",
		},
	}

	if err := partner.Connection.WriteJSON(rejectMsg); err != nil {
		log.Printf("Error sending call_rejected to partner %s: %v", partner.ID, err)
	}

	user.CallState = models.CallStateEnded
	partner.CallState = models.CallStateEnded

	log.Printf("Call rejected by %s from %s", user.ID, partner.ID)
}

func (s *SignalingServer) handleCallEnd(msg Message, user *models.User) {
	partner := s.UserPool.FindPartner(user.ID)
	if partner != nil {
		// Send call_ended to partner
		endMsg := Message{
			Type:      "call_ended",
			From:      user.ID,
			To:        partner.ID,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"reason": "Call ended by peer",
			},
		}

		if err := partner.Connection.WriteJSON(endMsg); err != nil {
			log.Printf("Error sending call_ended to partner %s: %v", partner.ID, err)
		}

		partner.CallState = models.CallStateEnded
	}

	user.CallState = models.CallStateEnded

	// Update room state
	if roomID := user.RoomID; roomID != "" {
		if room := s.UserPool.Rooms[roomID]; room != nil {
			room.CallState = models.CallStateEnded
			now := time.Now()
			room.EndedAt = &now
		}
	}

	log.Printf("Call ended by %s", user.ID)
}

func (s *SignalingServer) handleGetICEServers(user *models.User) {
	iceServers := s.GetICEServers()

	response := Message{
		Type:      "ice_servers",
		Timestamp: time.Now(),
		Payload:   iceServers,
	}

	if err := user.Connection.WriteJSON(response); err != nil {
		log.Printf("Error sending ICE servers to user %s: %v", user.ID, err)
	}
}

// Enhanced validation functions with detailed error reporting

func (s *SignalingServer) validateSDPOfferDetailed(payload interface{}) (bool, string) {
	data, ok := payload.(map[string]interface{})
	if !ok {
		return false, "payload must be a JSON object"
	}

	// Check for required fields
	sdp, sdpExists := data["sdp"]
	if !sdpExists {
		return false, "missing 'sdp' field in payload"
	}

	sdpStr, ok := sdp.(string)
	if !ok {
		return false, "SDP must be a string"
	}

	if len(sdpStr) == 0 {
		return false, "SDP cannot be empty"
	}

	// Check for type field
	if sdpType, exists := data["type"]; exists {
		if typeStr, ok := sdpType.(string); ok && typeStr != "offer" {
			return false, fmt.Sprintf("expected type 'offer', got '%s'", typeStr)
		}
	}

	// More flexible SDP validation - check for basic SDP structure
	if !strings.HasPrefix(sdpStr, "v=0") {
		return false, "SDP must start with 'v=0'"
	}

	// Check for essential SDP lines (more flexible than regex)
	requiredLines := []string{"o=", "s=", "t="}
	for _, line := range requiredLines {
		if !strings.Contains(sdpStr, line) {
			return false, fmt.Sprintf("SDP missing required line starting with '%s'", line)
		}
	}

	// Check for media line (audio for voice chat)
	if !strings.Contains(sdpStr, "m=audio") && !strings.Contains(sdpStr, "m=application") {
		return false, "SDP must contain at least one media line (m=audio or m=application)"
	}

	log.Printf("[DEBUG] SDP offer validation passed: %d characters, contains required elements", len(sdpStr))
	return true, ""
}

func (s *SignalingServer) validateSDPAnswerDetailed(payload interface{}) (bool, string) {
	data, ok := payload.(map[string]interface{})
	if !ok {
		return false, "payload must be a JSON object"
	}

	// Check for required fields
	sdp, sdpExists := data["sdp"]
	if !sdpExists {
		return false, "missing 'sdp' field in payload"
	}

	sdpStr, ok := sdp.(string)
	if !ok {
		return false, "SDP must be a string"
	}

	if len(sdpStr) == 0 {
		return false, "SDP cannot be empty"
	}

	// Check for type field
	if sdpType, exists := data["type"]; exists {
		if typeStr, ok := sdpType.(string); ok && typeStr != "answer" {
			return false, fmt.Sprintf("expected type 'answer', got '%s'", typeStr)
		}
	}

	// More flexible SDP validation - check for basic SDP structure
	if !strings.HasPrefix(sdpStr, "v=0") {
		return false, "SDP must start with 'v=0'"
	}

	// Check for essential SDP lines
	requiredLines := []string{"o=", "s=", "t="}
	for _, line := range requiredLines {
		if !strings.Contains(sdpStr, line) {
			return false, fmt.Sprintf("SDP missing required line starting with '%s'", line)
		}
	}

	// Check for media line
	if !strings.Contains(sdpStr, "m=audio") && !strings.Contains(sdpStr, "m=application") {
		return false, "SDP must contain at least one media line (m=audio or m=application)"
	}

	log.Printf("[DEBUG] SDP answer validation passed: %d characters, contains required elements", len(sdpStr))
	return true, ""
}

// Legacy validation functions (kept for backward compatibility, but improved)
func (s *SignalingServer) validateSDPOffer(payload interface{}) bool {
	valid, _ := s.validateSDPOfferDetailed(payload)
	return valid
}

func (s *SignalingServer) validateSDPAnswer(payload interface{}) bool {
	valid, _ := s.validateSDPAnswerDetailed(payload)
	return valid
}

func (s *SignalingServer) validateICECandidateDetailed(payload interface{}) (bool, string) {
	data, ok := payload.(map[string]interface{})
	if !ok {
		return false, "payload must be a JSON object"
	}

	// Check for required candidate field
	candidate, candidateExists := data["candidate"]
	if !candidateExists {
		return false, "missing 'candidate' field in payload"
	}

	candidateStr, ok := candidate.(string)
	if !ok {
		return false, "candidate must be a string"
	}

	if len(candidateStr) == 0 {
		return false, "candidate cannot be empty"
	}

	// Basic ICE candidate validation
	if !strings.Contains(candidateStr, "candidate:") {
		return false, "candidate must contain 'candidate:' prefix"
	}

	// Check for optional but common fields
	if sdpMLineIndex, exists := data["sdpMLineIndex"]; exists {
		if _, ok := sdpMLineIndex.(float64); !ok {
			// JSON numbers are parsed as float64 in Go
			return false, "sdpMLineIndex must be a number"
		}
	}

	if sdpMid, exists := data["sdpMid"]; exists {
		if _, ok := sdpMid.(string); !ok {
			return false, "sdpMid must be a string"
		}
	}

	log.Printf("[DEBUG] ICE candidate validation passed: %s", candidateStr)
	return true, ""
}

func (s *SignalingServer) validateICECandidate(payload interface{}) bool {
	valid, _ := s.validateICECandidateDetailed(payload)
	return valid
}

// Utility functions

func (s *SignalingServer) sendError(user *models.User, message string) {
	errorMsg := Message{
		Type:      "error",
		Timestamp: time.Now(),
		Payload: map[string]string{
			"message": message,
		},
	}

	if err := user.Connection.WriteJSON(errorMsg); err != nil {
		log.Printf("Error sending error message to user %s: %v", user.ID, err)
	}
}

func (s *SignalingServer) GetICEServers() ICEServersResponse {
	var iceServers []ICEServer

	// Add STUN servers
	if len(s.STUNServers) > 0 {
		iceServers = append(iceServers, ICEServer{
			URLs: s.STUNServers,
		})
	} else {
		// Default public STUN servers
		iceServers = append(iceServers, ICEServer{
			URLs: []string{
				"stun:stun.l.google.com:19302",
				"stun:stun1.l.google.com:19302",
				"stun:stun2.l.google.com:19302",
			},
		})
	}

	// Add TURN servers
	for _, turnServer := range s.TURNServers {
		iceServers = append(iceServers, ICEServer{
			URLs:       []string{turnServer.URL},
			Username:   turnServer.Username,
			Credential: turnServer.Credential,
		})
	}

	return ICEServersResponse{
		ICEServers: iceServers,
	}
}

// GetStats returns current server statistics
func (s *SignalingServer) GetStats() map[string]interface{} {
	stats := s.UserPool.GetStats()
	return map[string]interface{}{
		"waiting_users": stats["waiting_users"],
		"active_users":  stats["active_users"],
		"active_rooms":  stats["active_rooms"],
		"server_uptime": time.Now().Format(time.RFC3339),
	}
}
