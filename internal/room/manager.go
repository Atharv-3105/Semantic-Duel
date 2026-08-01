package room

import (
	"sync"

	"github.com/Atharv-3105/Semantic-Duel/internal/logger"
	"github.com/Atharv-3105/Semantic-Duel/internal/metrics"
	"github.com/Atharv-3105/Semantic-Duel/internal/ws"
)

type Manager struct {
	rooms        map[string]*Room
	clientToRoom map[string]*Room
	mu           sync.RWMutex
	log          *logger.Logger
}

func NewManager(log *logger.Logger) *Manager {
	return &Manager{
		rooms:        make(map[string]*Room),
		clientToRoom: make(map[string]*Room),
		log:          log,
	}
}

func (m *Manager) Add(room *Room) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rooms[room.ID] = room
	m.clientToRoom[room.Player1.ID] = room
	m.clientToRoom[room.Player2.ID] = room
	m.log.Info("[ROOM] room added", "room_id", room.ID, "total_rooms", len(m.rooms))
}

func (m *Manager) RoomForClient(clientID string) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, ok := m.clientToRoom[clientID]
	return room, ok
}

// HandleDisconnect forces the game to end with the disconnecting player as the loser.
// Cleanup (removing the room from the maps) is handled exclusively by the cleanupCh
// consumer in main.go, which calls CleanupRoom after receiving the room ID.
// This avoids the double-cleanup bug that occurred when CleanupRoom was called here
// directly AND again from the cleanupCh consumer.
func (m *Manager) HandleDisconnect(clientID string) {
	room, ok := m.RoomForClient(clientID)
	if !ok {
		return
	}

	m.log.Println("[ROOM] disconnect detected, ending game for:", clientID)
	metrics.IncDisconnects()
	room.ForceEnd(clientID)
	// Do NOT call m.CleanupRoom here.
	// ForceEnd → room.cleanup() → onCleanup(roomID) → cleanupCh → main.go → CleanupRoom.
}

// CleanupRoom removes the room and both player mappings from the manager.
// Returns both players so the caller can re-enqueue them for a new match.
func (m *Manager) CleanupRoom(roomID string) (*ws.Client, *ws.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[roomID]
	if !ok {
		return nil, nil
	}

	p1 := room.Player1
	p2 := room.Player2

	delete(m.clientToRoom, p1.ID)
	delete(m.clientToRoom, p2.ID)
	delete(m.rooms, roomID)

	m.log.Println("[ROOM] cleaned up:", roomID)
	return p1, p2
}
