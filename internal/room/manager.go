package room

import (
	"sync"
	// "time"
	"github.com/Atharv-3105/Graph-Duel/internal/logger"
	"github.com/Atharv-3105/Graph-Duel/internal/metrics"
	"github.com/Atharv-3105/Graph-Duel/internal/ws"
)

type Manager struct {
	rooms 	map[string]*Room
	clientToRoom map[string]*Room
	mu    	sync.RWMutex
	log 	*logger.Logger
}


func NewManager(log *logger.Logger) *Manager{
	return &Manager{
		rooms:	make(map[string]*Room),
		log:	log,
		clientToRoom: make(map[string]*Room),
	}
}



func (m *Manager) Add(room *Room) {
	m.mu.Lock()
	defer m.mu.Unlock()


	m.rooms[room.ID] = room

	m.clientToRoom[room.Player1.ID] = room
	m.clientToRoom[room.Player2.ID] = room
	m.log.Info("[ROOM] room added", "room_id", room.ID, "total rooms", len(m.rooms))
}

func (m *Manager) Remove(roomID string){
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rooms, roomID)
}

func (m *Manager) RoomForClient(clientID string) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, ok := m.clientToRoom[clientID]
	return room, ok
}


func (m *Manager) HandleDisconnect(clientID string) {
	room, ok := m.RoomForClient(clientID)
	if !ok {
		return
	}

	m.log.Println("[ROOM] disconnect detected, forcing game end: ", clientID)
	metrics.IncDisconnects()
	room.ForceEnd(clientID)
	m.CleanupRoom(room.ID)
}

func (m *Manager) CleanupRoom(roomID string) (*ws.Client, *ws.Client){
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

	m.log.Println("[ROOM] cleaned up: ", roomID)
	return p1, p2
}


