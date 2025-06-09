package models

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebRTC-specific message types
type SDPMessage struct {
	Type string `json:"type"` // "offer" or "answer"
	SDP  string `json:"sdp"`
}

type ICECandidateMessage struct {
	Candidate     string `json:"candidate"`
	SDPMLineIndex int    `json:"sdpMLineIndex"`
	SDPMid        string `json:"sdpMid"`
}

type CallState string

type Connection struct {
	Conn     *websocket.Conn
	UserID   string
	LastPing time.Time
	IsActive bool
	mutex    sync.RWMutex
}

func (c *Connection) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.IsActive = false
	return c.Conn.Close()
}

func (c *Connection) WriteJSON(v interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.IsActive {
		return websocket.ErrCloseSent
	}
	return c.Conn.WriteJSON(v)
}

func (c *Connection) UpdatePing() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.LastPing = time.Now()
}

type User struct {
	ID          string      `json:"id"`
	SessionID   string      `json:"session_id"`
	Status      string      `json:"status"` // waiting, matched, connected, disconnected
	ConnectedAt time.Time   `json:"connected_at"`
	Connection  *Connection `json:"-"` // Don't serialize connection
	PartnerID   string      `json:"partner_id,omitempty"`
	RoomID      string      `json:"room_id,omitempty"`
	CallState   CallState   `json:"call_state"`
	MediaInfo   *MediaInfo  `json:"media_info,omitempty"`
}

type MediaInfo struct {
	HasAudio bool   `json:"has_audio"`
	HasVideo bool   `json:"has_video"`
	Codec    string `json:"codec,omitempty"`
}

type Room struct {
	ID        string     `json:"id"`
	User1ID   string     `json:"user1_id"`
	User2ID   string     `json:"user2_id"`
	CreatedAt time.Time  `json:"created_at"`
	IsActive  bool       `json:"is_active"`
	CallState CallState  `json:"call_state"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

type UserPool struct {
	WaitingUsers map[string]*User
	ActiveUsers  map[string]*User
	Rooms        map[string]*Room
	UserRooms    map[string]string // userID -> roomID mapping
	mutex        sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewUserPool() *UserPool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &UserPool{
		WaitingUsers: make(map[string]*User),
		ActiveUsers:  make(map[string]*User),
		Rooms:        make(map[string]*Room),
		UserRooms:    make(map[string]string),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Start cleanup goroutine
	go pool.cleanupInactiveConnections()
	return pool
}

func (p *UserPool) AddWaitingUser(user *User) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	user.Status = StatusWaiting
	user.ConnectedAt = time.Now()
	p.WaitingUsers[user.ID] = user
}

func (p *UserPool) GetRandomWaitingUser(excludeID string) *User {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	for id, user := range p.WaitingUsers {
		if id != excludeID {
			return user
		}
	}
	return nil
}

func (p *UserPool) CreateRoom(user1 *User, user2 *User) *Room {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	roomID := generateRoomID()
	room := &Room{
		ID:        roomID,
		User1ID:   user1.ID,
		User2ID:   user2.ID,
		CreatedAt: time.Now(),
		IsActive:  true,
		CallState: CallState(CallStateIdle),
	}

	// Update users
	user1.Status = StatusConnected
	user1.PartnerID = user2.ID
	user1.RoomID = roomID
	user1.CallState = CallState(CallStateIdle)

	user2.Status = StatusConnected
	user2.PartnerID = user1.ID
	user2.RoomID = roomID
	user2.CallState = CallState(CallStateIdle)

	// Move users to active and create room mappings
	delete(p.WaitingUsers, user1.ID)
	delete(p.WaitingUsers, user2.ID)
	p.ActiveUsers[user1.ID] = user1
	p.ActiveUsers[user2.ID] = user2
	p.Rooms[roomID] = room
	p.UserRooms[user1.ID] = roomID
	p.UserRooms[user2.ID] = roomID

	return room
}

func (p *UserPool) MoveToActive(userID string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if user, exists := p.WaitingUsers[userID]; exists {
		delete(p.WaitingUsers, userID)
		p.ActiveUsers[userID] = user
		user.Status = StatusConnected
	}
}

func (p *UserPool) GetActiveUser(userID string) *User {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.ActiveUsers[userID]
}

func (p *UserPool) GetUser(userID string) *User {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if user := p.WaitingUsers[userID]; user != nil {
		return user
	}
	return p.ActiveUsers[userID]
}

func (p *UserPool) RemoveUser(userID string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Clean up room if user was in one
	if roomID, exists := p.UserRooms[userID]; exists {
		if room := p.Rooms[roomID]; room != nil {
			room.IsActive = false
			// Remove partner's room mapping too
			partnerID := ""
			if room.User1ID == userID {
				partnerID = room.User2ID
			} else {
				partnerID = room.User1ID
			}
			delete(p.UserRooms, partnerID)
		}
		delete(p.UserRooms, userID)
	}

	delete(p.WaitingUsers, userID)
	delete(p.ActiveUsers, userID)
}

func (p *UserPool) FindPartner(userID string) *User {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if roomID, exists := p.UserRooms[userID]; exists {
		if room := p.Rooms[roomID]; room != nil && room.IsActive {
			partnerID := ""
			if room.User1ID == userID {
				partnerID = room.User2ID
			} else {
				partnerID = room.User1ID
			}
			return p.ActiveUsers[partnerID]
		}
	}
	return nil
}

func (p *UserPool) MoveToWaiting(userID string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if user, exists := p.ActiveUsers[userID]; exists {
		delete(p.ActiveUsers, userID)
		p.WaitingUsers[userID] = user
		user.Status = "waiting"
		user.PartnerID = ""
		user.RoomID = ""
	}
}

func (p *UserPool) GetStats() map[string]int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return map[string]int{
		"waiting_users": len(p.WaitingUsers),
		"active_users":  len(p.ActiveUsers),
		"active_rooms":  len(p.Rooms),
	}
}

// Cleanup inactive connections periodically
func (p *UserPool) cleanupInactiveConnections() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.performCleanup()
		}
	}
}

func (p *UserPool) performCleanup() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	cutoff := time.Now().Add(-5 * time.Minute)

	// Clean up waiting users with old connections
	for id, user := range p.WaitingUsers {
		if user.Connection != nil && user.Connection.LastPing.Before(cutoff) {
			delete(p.WaitingUsers, id)
			user.Connection.Close()
		}
	}

	// Clean up active users with old connections
	for id, user := range p.ActiveUsers {
		if user.Connection != nil && user.Connection.LastPing.Before(cutoff) {
			delete(p.ActiveUsers, id)
			user.Connection.Close()
			// Also clean up room
			if roomID := p.UserRooms[id]; roomID != "" {
				if room := p.Rooms[roomID]; room != nil {
					room.IsActive = false
				}
				delete(p.UserRooms, id)
			}
		}
	}
}

func (p *UserPool) Shutdown() {
	p.cancel()
}

func generateRoomID() string {
	// Simple room ID generation - in production, use UUID
	return time.Now().Format("20060102150405") + "-room"
}
