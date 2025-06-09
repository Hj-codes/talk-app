package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"voice-chat-app/handlers"
	"voice-chat-app/models"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration test setup
func setupTestServer() (*httptest.Server, *handlers.SignalingServer) {
	userPool := models.NewUserPool()
	signalingServer := &handlers.SignalingServer{
		UserPool: userPool,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", signalingServer.HandleWebSocket)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(signalingServer.GetStats())
	})

	server := httptest.NewServer(mux)
	return server, signalingServer
}

func connectWebSocket(t testing.TB, serverURL string) (*websocket.Conn, handlers.Message) {
	wsURL := "ws" + strings.TrimPrefix(serverURL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL+"/ws", nil)
	require.NoError(t, err)

	// Read session message
	var sessionMsg handlers.Message
	err = conn.ReadJSON(&sessionMsg)
	require.NoError(t, err)
	require.Equal(t, "session", sessionMsg.Type)

	return conn, sessionMsg
}

func TestIntegration_SingleUserConnection(t *testing.T) {
	server, signalingServer := setupTestServer()
	defer server.Close()
	defer signalingServer.UserPool.Shutdown()

	// Connect a single user
	conn, sessionMsg := connectWebSocket(t, server.URL)
	defer conn.Close()

	// Verify session message
	payload, ok := sessionMsg.Payload.(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, payload["user_id"])
	assert.NotEmpty(t, payload["token"])

	// Check server stats
	stats := signalingServer.GetStats()

	// Convert interface{} to expected types and check values
	waitingUsers, ok := stats["waiting_users"].(int)
	require.True(t, ok, "waiting_users should be an int")
	assert.Equal(t, 1, waitingUsers)

	activeUsers, ok := stats["active_users"].(int)
	require.True(t, ok, "active_users should be an int")
	assert.Equal(t, 0, activeUsers)

	activeRooms, ok := stats["active_rooms"].(int)
	require.True(t, ok, "active_rooms should be an int")
	assert.Equal(t, 0, activeRooms)
}

func TestIntegration_TwoUserMatchmaking(t *testing.T) {
	server, signalingServer := setupTestServer()
	defer server.Close()
	defer signalingServer.UserPool.Shutdown()

	// Connect first user
	conn1, sessionMsg1 := connectWebSocket(t, server.URL)
	defer conn1.Close()

	// Connect second user
	conn2, sessionMsg2 := connectWebSocket(t, server.URL)
	defer conn2.Close()

	// Extract user IDs
	payload1 := sessionMsg1.Payload.(map[string]interface{})
	payload2 := sessionMsg2.Payload.(map[string]interface{})
	userID1 := payload1["user_id"].(string)
	userID2 := payload2["user_id"].(string)

	// User 1 requests a match
	findMatchMsg := handlers.Message{
		Type: "find_match",
		From: userID1,
	}
	err := conn1.WriteJSON(findMatchMsg)
	require.NoError(t, err)

	// Both users should receive match_found messages
	var matchMsg1, matchMsg2 handlers.Message

	// Read from conn1
	err = conn1.ReadJSON(&matchMsg1)
	require.NoError(t, err)
	assert.Equal(t, "match_found", matchMsg1.Type)

	// Read from conn2
	err = conn2.ReadJSON(&matchMsg2)
	require.NoError(t, err)
	assert.Equal(t, "match_found", matchMsg2.Type)

	// Verify match details
	matchPayload1 := matchMsg1.Payload.(map[string]interface{})
	matchPayload2 := matchMsg2.Payload.(map[string]interface{})

	assert.Equal(t, userID2, matchPayload1["partner_id"])
	assert.Equal(t, userID1, matchPayload2["partner_id"])
	assert.Equal(t, matchPayload1["room_id"], matchPayload2["room_id"])

	// Check server stats after matching
	stats := signalingServer.GetStats()

	// Convert interface{} to expected types and check values
	waitingUsers, ok := stats["waiting_users"].(int)
	require.True(t, ok, "waiting_users should be an int")
	assert.Equal(t, 0, waitingUsers)

	activeUsers, ok := stats["active_users"].(int)
	require.True(t, ok, "active_users should be an int")
	assert.Equal(t, 2, activeUsers)

	activeRooms, ok := stats["active_rooms"].(int)
	require.True(t, ok, "active_rooms should be an int")
	assert.Equal(t, 1, activeRooms)
}

func TestIntegration_HealthAndStatsEndpoints(t *testing.T) {
	server, signalingServer := setupTestServer()
	defer server.Close()
	defer signalingServer.UserPool.Shutdown()

	// Test health endpoint
	resp, err := http.Get(server.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var healthResponse map[string]string
	err = json.NewDecoder(resp.Body).Decode(&healthResponse)
	require.NoError(t, err)
	assert.Equal(t, "healthy", healthResponse["status"])

	// Test stats endpoint
	resp, err = http.Get(server.URL + "/stats")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var statsResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&statsResponse)
	require.NoError(t, err)

	// Should have expected stat fields
	assert.Contains(t, statsResponse, "waiting_users")
	assert.Contains(t, statsResponse, "active_users")
	assert.Contains(t, statsResponse, "active_rooms")
	assert.Contains(t, statsResponse, "server_uptime")
}
