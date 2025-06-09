package handlers

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"
	"voice-chat-app/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignalingServer_GetStats(t *testing.T) {
	userPool := models.NewUserPool()
	defer userPool.Shutdown()

	server := &SignalingServer{
		UserPool: userPool,
	}

	// Add some test users (without WebSocket connections for testing)
	conn1 := &models.Connection{UserID: "user1", IsActive: true}
	conn2 := &models.Connection{UserID: "user2", IsActive: true}
	conn3 := &models.Connection{UserID: "user3", IsActive: true}
	conn4 := &models.Connection{UserID: "user4", IsActive: true}

	user1 := &models.User{ID: "user1", Connection: conn1}
	user2 := &models.User{ID: "user2", Connection: conn2}
	user3 := &models.User{ID: "user3", Connection: conn3}
	user4 := &models.User{ID: "user4", Connection: conn4}

	userPool.AddWaitingUser(user1)
	userPool.AddWaitingUser(user2)
	userPool.AddWaitingUser(user3)
	userPool.AddWaitingUser(user4)

	// Create a room
	userPool.CreateRoom(user3, user4)

	stats := server.GetStats()

	assert.Equal(t, 2, stats["waiting_users"])
	assert.Equal(t, 2, stats["active_users"])
	assert.Equal(t, 1, stats["active_rooms"])
	assert.NotNil(t, stats["server_uptime"])
}

func TestSignalingServer_UserPoolOperations(t *testing.T) {
	userPool := models.NewUserPool()
	defer userPool.Shutdown()

	// Test adding users to the pool
	conn1 := &models.Connection{UserID: "user1", IsActive: true}
	conn2 := &models.Connection{UserID: "user2", IsActive: true}

	user1 := &models.User{ID: "user1", Connection: conn1}
	user2 := &models.User{ID: "user2", Connection: conn2}

	userPool.AddWaitingUser(user1)
	userPool.AddWaitingUser(user2)

	// Verify users are in waiting state
	assert.Equal(t, 2, len(userPool.WaitingUsers))
	assert.Equal(t, 0, len(userPool.ActiveUsers))

	// Test room creation
	room := userPool.CreateRoom(user1, user2)

	// Verify room creation and user state changes
	assert.NotNil(t, room)
	assert.Equal(t, 0, len(userPool.WaitingUsers))
	assert.Equal(t, 2, len(userPool.ActiveUsers))
	assert.Equal(t, "connected", user1.Status)
	assert.Equal(t, "connected", user2.Status)
	assert.Equal(t, user2.ID, user1.PartnerID)
	assert.Equal(t, user1.ID, user2.PartnerID)
}

func TestSignalingServer_MatchmakingLogic(t *testing.T) {
	userPool := models.NewUserPool()
	defer userPool.Shutdown()

	// Test case 1: No available partner
	conn1 := &models.Connection{UserID: "lonely-user", IsActive: true}
	user1 := &models.User{ID: "lonely-user", Connection: conn1}
	userPool.AddWaitingUser(user1)

	// Check that user remains in waiting state when no partner is available
	partner := userPool.GetRandomWaitingUser(user1.ID)
	assert.Nil(t, partner)
	assert.Equal(t, "waiting", user1.Status)

	// Test case 2: Partner available
	conn2 := &models.Connection{UserID: "user2", IsActive: true}
	user2 := &models.User{ID: "user2", Connection: conn2}
	userPool.AddWaitingUser(user2)

	// Now user1 should be able to find user2 as a partner
	partner = userPool.GetRandomWaitingUser(user1.ID)
	assert.NotNil(t, partner)
	assert.Equal(t, user2.ID, partner.ID)

	// Create room for both users
	room := userPool.CreateRoom(user1, user2)
	assert.NotNil(t, room)
	assert.Equal(t, user1.RoomID, user2.RoomID)
}

func TestSignalingServer_ConcurrentOperations(t *testing.T) {
	userPool := models.NewUserPool()
	defer userPool.Shutdown()

	var server *SignalingServer

	numUsers := 50
	var wg sync.WaitGroup

	// Add users concurrently
	wg.Add(numUsers)
	for i := 0; i < numUsers; i++ {
		go func(userID int) {
			defer wg.Done()

			conn := &models.Connection{
				UserID:   fmt.Sprintf("user-%d", userID),
				IsActive: true,
			}
			user := &models.User{
				ID:         fmt.Sprintf("user-%d", userID),
				Connection: conn,
			}

			userPool.AddWaitingUser(user)
		}(i)
	}

	wg.Wait()

	// Verify all users were added
	assert.Equal(t, numUsers, len(userPool.WaitingUsers))

	// Test concurrent matchmaking
	wg.Add(numUsers / 2)
	for i := 0; i < numUsers; i += 2 {
		go func(userIndex int) {
			defer wg.Done()

			user1ID := fmt.Sprintf("user-%d", userIndex)
			user2ID := fmt.Sprintf("user-%d", userIndex+1)

			user1 := userPool.GetUser(user1ID)
			user2 := userPool.GetUser(user2ID)

			if user1 != nil && user2 != nil {
				userPool.CreateRoom(user1, user2)
			}
		}(i)
	}

	wg.Wait()

	// Verify system state is consistent
	server = &SignalingServer{UserPool: userPool}
	stats := server.GetStats()
	waitingUsers := stats["waiting_users"].(int)
	activeUsers := stats["active_users"].(int)
	activeRooms := stats["active_rooms"].(int)

	// Total users should be consistent
	totalUsers := waitingUsers + activeUsers
	assert.Equal(t, numUsers, totalUsers)

	// Active users should be even (paired)
	assert.Equal(t, 0, activeUsers%2)

	// Number of rooms should match active pairs (but may be less due to race conditions)
	assert.True(t, activeRooms <= activeUsers/2)
}

func TestMessage_Serialization(t *testing.T) {
	msg := Message{
		Type:      "test",
		Payload:   map[string]string{"key": "value"},
		From:      "user1",
		To:        "user2",
		Timestamp: time.Now(),
	}

	// Test JSON marshaling
	data, err := json.Marshal(msg)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled Message
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, msg.Type, unmarshaled.Type)
	assert.Equal(t, msg.From, unmarshaled.From)
	assert.Equal(t, msg.To, unmarshaled.To)
}

func TestSignalingServer_EdgeCases(t *testing.T) {
	userPool := models.NewUserPool()
	defer userPool.Shutdown()

	t.Run("empty user pool stats", func(t *testing.T) {
		server := &SignalingServer{UserPool: userPool}
		stats := server.GetStats()
		assert.Equal(t, 0, stats["waiting_users"])
		assert.Equal(t, 0, stats["active_users"])
		assert.Equal(t, 0, stats["active_rooms"])
	})

	t.Run("user removal", func(t *testing.T) {
		conn := &models.Connection{UserID: "test-user", IsActive: true}
		user := &models.User{ID: "test-user", Connection: conn}

		userPool.AddWaitingUser(user)
		assert.Equal(t, 1, len(userPool.WaitingUsers))

		userPool.RemoveUser(user.ID)
		assert.Equal(t, 0, len(userPool.WaitingUsers))
	})

	t.Run("room cleanup on user removal", func(t *testing.T) {
		conn1 := &models.Connection{UserID: "user1", IsActive: true}
		conn2 := &models.Connection{UserID: "user2", IsActive: true}

		user1 := &models.User{ID: "user1", Connection: conn1}
		user2 := &models.User{ID: "user2", Connection: conn2}

		userPool.AddWaitingUser(user1)
		userPool.AddWaitingUser(user2)
		room := userPool.CreateRoom(user1, user2)

		assert.True(t, room.IsActive)
		assert.Equal(t, 1, len(userPool.Rooms))

		// Remove one user
		userPool.RemoveUser(user1.ID)

		// Room should be deactivated
		assert.False(t, userPool.Rooms[room.ID].IsActive)
	})
}

// Benchmark tests
func BenchmarkSignalingServer_UserOperations(b *testing.B) {
	userPool := models.NewUserPool()
	defer userPool.Shutdown()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		userCounter := 0
		for pb.Next() {
			conn := &models.Connection{
				UserID:   fmt.Sprintf("bench-user-%d", userCounter),
				IsActive: true,
			}
			user := &models.User{
				ID:         fmt.Sprintf("bench-user-%d", userCounter),
				Connection: conn,
			}
			userCounter++

			userPool.AddWaitingUser(user)

			// Try to find a partner
			partner := userPool.GetRandomWaitingUser(user.ID)
			if partner != nil {
				userPool.CreateRoom(user, partner)
			}
		}
	})
}
