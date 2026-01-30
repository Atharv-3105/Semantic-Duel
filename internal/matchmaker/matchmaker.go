package matchmaker

import (
	"fmt"
	// "log"
	"sync"
	"time"

	"github.com/Atharv-3105/Graph-Duel/internal/logger"
	"github.com/Atharv-3105/Graph-Duel/internal/room"
	"github.com/Atharv-3105/Graph-Duel/internal/semantic"
	"github.com/Atharv-3105/Graph-Duel/internal/ws"
)


type Matchmaker struct{
	queue	[]*ws.Client
	mu		sync.Mutex
	rm 		*room.Manager
	log 	*logger.Logger
	semantic *semantic.Client
	cleanupCh chan<- string
	gameDuration 	int 
	rateLimitSeconds int 
}


func New(rm *room.Manager, log *logger.Logger, sc *semantic.Client, cleanupCh chan<- string, gameDuration int, rateLimitSeconds int,) *Matchmaker{
	return &Matchmaker{
		queue:	 make([]*ws.Client, 0),
		rm:		 rm,
		log:	log,
		semantic: 	sc,
		cleanupCh: cleanupCh,
		gameDuration: gameDuration,
		rateLimitSeconds: rateLimitSeconds,
	}
}


func (m *Matchmaker) Enqueue(client *ws.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.queue = append(m.queue, client)

	if len(m.queue) >= 2 {
		p1 := m.queue[0]
		p2 := m.queue[1]
		m.queue = m.queue[2:]
		m.log.Info("[MATCH] pairing players", "queue_size", len(m.queue))

		roomID := fmt.Sprintf("room-%d", time.Now().UnixNano())
		
		onCleanup := func(roomID string) {
			m.cleanupCh <- roomID
		}

		room := room.NewRoom(roomID, p1, p2, m.semantic, onCleanup, m.gameDuration, m.rateLimitSeconds)
		room.Start("FIRE")
		m.rm.Add(room)
		m.log.Info("[MATCH] room created", "room_id", roomID)
	}
}