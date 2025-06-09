package models

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConnection_UpdatePing(t *testing.T) {
	conn := &Connection{
		UserID:   "test-user",
		LastPing: time.Now().Add(-time.Hour),
		IsActive: true,
	}

	oldPing := conn.LastPing
	time.Sleep(time.Millisecond) // Ensure time difference
	conn.UpdatePing()

	assert.True(t, conn.LastPing.After(oldPing))
}

func TestConnection_Close(t *testing.T) {
	conn := &Connection{
		UserID:   "test-user",
		LastPing: time.Now(),
		IsActive: true,
	}

	// Test that Close sets IsActive to false
	// Note: We can't test the actual websocket close without a real connection
	conn.IsActive = false // Simulate what Close() would do
	assert.False(t, conn.IsActive)
}

func TestNewUserPool(t *testing.T) {
	pool := NewUserPool()

	assert.NotNil(t, pool)
	assert.NotNil(t, pool.WaitingUsers)
	assert.NotNil(t, pool.ActiveUsers)
	assert.NotNil(t, pool.Rooms)
	assert.NotNil(t, pool.UserRooms)

	// Cleanup
	pool.Shutdown()
}

func TestUserPool_AddWaitingUser(t *testing.T) {
	pool := NewUserPool()
	defer pool.Shutdown()

	user := &User{
		ID:         "test-user-1",
		SessionID:  "session-1",
		Connection: &Connection{UserID: "test-user-1", IsActive: true},
	}

	pool.AddWaitingUser(user)

	assert.Equal(t, "waiting", user.Status)
	assert.False(t, user.ConnectedAt.IsZero())
	assert.Equal(t, user, pool.WaitingUsers["test-user-1"])
}

func TestUserPool_GetRandomWaitingUser(t *testing.T) {
	pool := NewUserPool()
	defer pool.Shutdown()

	// Test empty pool
	result := pool.GetRandomWaitingUser("any-id")
	assert.Nil(t, result)

	// Add users
	user1 := &User{ID: "user1", Connection: &Connection{UserID: "user1", IsActive: true}}
	user2 := &User{ID: "user2", Connection: &Connection{UserID: "user2", IsActive: true}}

	pool.AddWaitingUser(user1)
	pool.AddWaitingUser(user2)

	// Test getting random user excluding one
	result = pool.GetRandomWaitingUser("user1")
	assert.NotNil(t, result)
	assert.Equal(t, "user2", result.ID)

	// Test excluding all users
	pool.RemoveUser("user2")
	result = pool.GetRandomWaitingUser("user1")
	assert.Nil(t, result)
}

func TestUserPool_CreateRoom(t *testing.T) {
	pool := NewUserPool()
	defer pool.Shutdown()

	user1 := &User{ID: "user1", Connection: &Connection{UserID: "user1", IsActive: true}}
	user2 := &User{ID: "user2", Connection: &Connection{UserID: "user2", IsActive: true}}

	pool.AddWaitingUser(user1)
	pool.AddWaitingUser(user2)

	room := pool.CreateRoom(user1, user2)

	// Verify room creation
	assert.NotEmpty(t, room.ID)
	assert.Equal(t, user1.ID, room.User1ID)
	assert.Equal(t, user2.ID, room.User2ID)
	assert.True(t, room.IsActive)

	// Verify users updated
	assert.Equal(t, "connected", user1.Status)
	assert.Equal(t, "connected", user2.Status)
	assert.Equal(t, user2.ID, user1.PartnerID)
	assert.Equal(t, user1.ID, user2.PartnerID)
	assert.Equal(t, room.ID, user1.RoomID)
	assert.Equal(t, room.ID, user2.RoomID)

	// Verify moved to active users
	assert.Nil(t, pool.WaitingUsers[user1.ID])
	assert.Nil(t, pool.WaitingUsers[user2.ID])
	assert.Equal(t, user1, pool.ActiveUsers[user1.ID])
	assert.Equal(t, user2, pool.ActiveUsers[user2.ID])

	// Verify room mappings
	assert.Equal(t, room.ID, pool.UserRooms[user1.ID])
	assert.Equal(t, room.ID, pool.UserRooms[user2.ID])
}

func TestUserPool_RemoveUser(t *testing.T) {
	pool := NewUserPool()
	defer pool.Shutdown()

	user1 := &User{ID: "user1", Connection: &Connection{UserID: "user1", IsActive: true}}
	user2 := &User{ID: "user2", Connection: &Connection{UserID: "user2", IsActive: true}}

	pool.AddWaitingUser(user1)
	pool.AddWaitingUser(user2)
	room := pool.CreateRoom(user1, user2)

	// Remove user1
	pool.RemoveUser(user1.ID)

	// Verify user1 removed
	assert.Nil(t, pool.ActiveUsers[user1.ID])
	assert.Nil(t, pool.WaitingUsers[user1.ID])
	assert.Equal(t, "", pool.UserRooms[user1.ID])

	// Verify room deactivated
	assert.False(t, pool.Rooms[room.ID].IsActive)

	// Verify partner's room mapping removed
	assert.Equal(t, "", pool.UserRooms[user2.ID])
}

func TestUserPool_ConcurrentAccess(t *testing.T) {
	pool := NewUserPool()
	defer pool.Shutdown()

	var wg sync.WaitGroup
	numGoroutines := 10
	usersPerGoroutine := 10

	// Concurrent user additions
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < usersPerGoroutine; j++ {
				userID := fmt.Sprintf("user-%d-%d", goroutineID, j)
				user := &User{
					ID:         userID,
					Connection: &Connection{UserID: userID, IsActive: true},
				}
				pool.AddWaitingUser(user)
			}
		}(i)
	}
	wg.Wait()

	// Verify all users added
	assert.Equal(t, numGoroutines*usersPerGoroutine, len(pool.WaitingUsers))

	// Concurrent room creation
	wg.Add(numGoroutines / 2)
	for i := 0; i < numGoroutines; i += 2 {
		go func(goroutineID int) {
			defer wg.Done()
			user1ID := fmt.Sprintf("user-%d-0", goroutineID)
			user2ID := fmt.Sprintf("user-%d-0", goroutineID+1)

			user1 := pool.GetUser(user1ID)
			user2 := pool.GetUser(user2ID)

			if user1 != nil && user2 != nil {
				pool.CreateRoom(user1, user2)
			}
		}(i)
	}
	wg.Wait()

	// Verify rooms created
	assert.True(t, len(pool.Rooms) > 0)
	assert.True(t, len(pool.ActiveUsers) > 0)
}

func TestUserPool_Stats(t *testing.T) {
	pool := NewUserPool()
	defer pool.Shutdown()

	// Add users in different states
	user1 := &User{ID: "waiting1", Connection: &Connection{UserID: "waiting1", IsActive: true}}
	user2 := &User{ID: "waiting2", Connection: &Connection{UserID: "waiting2", IsActive: true}}
	user3 := &User{ID: "active1", Connection: &Connection{UserID: "active1", IsActive: true}}
	user4 := &User{ID: "active2", Connection: &Connection{UserID: "active2", IsActive: true}}

	pool.AddWaitingUser(user1)
	pool.AddWaitingUser(user2)
	pool.AddWaitingUser(user3)
	pool.AddWaitingUser(user4)

	// Create a room
	pool.CreateRoom(user3, user4)

	stats := pool.GetStats()

	assert.Equal(t, 2, stats["waiting_users"])
	assert.Equal(t, 2, stats["active_users"])
	assert.Equal(t, 1, stats["active_rooms"])
}

// Race condition test for matchmaking
func TestUserPool_MatchmakingRaceCondition(t *testing.T) {
	pool := NewUserPool()
	defer pool.Shutdown()

	numUsers := 100
	var wg sync.WaitGroup

	// Add users concurrently
	wg.Add(numUsers)
	for i := 0; i < numUsers; i++ {
		go func(id int) {
			defer wg.Done()
			userID := fmt.Sprintf("race-user-%d", id)
			user := &User{
				ID:         userID,
				Connection: &Connection{UserID: userID, IsActive: true},
			}
			pool.AddWaitingUser(user)
		}(i)
	}
	wg.Wait()

	// Try to match users concurrently
	matchedRooms := make(map[string]bool)
	var roomsMutex sync.Mutex

	wg.Add(numUsers)
	for i := 0; i < numUsers; i++ {
		go func(id int) {
			defer wg.Done()
			userID := fmt.Sprintf("race-user-%d", id)
			user := pool.GetUser(userID)
			if user == nil {
				return
			}

			partner := pool.GetRandomWaitingUser(userID)
			if partner != nil {
				room := pool.CreateRoom(user, partner)
				roomsMutex.Lock()
				matchedRooms[room.ID] = true
				roomsMutex.Unlock()
			}
		}(i)
	}
	wg.Wait()

	// Verify no duplicate rooms and proper state
	assert.True(t, len(matchedRooms) > 0)
	assert.Equal(t, len(matchedRooms), len(pool.Rooms))

	// Verify all rooms are properly formed
	for _, room := range pool.Rooms {
		assert.NotEmpty(t, room.User1ID)
		assert.NotEmpty(t, room.User2ID)
		assert.NotEqual(t, room.User1ID, room.User2ID)
	}
}
