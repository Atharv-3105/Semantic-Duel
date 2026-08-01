package matchmaker

import (
	"fmt"
	"sync"
	"time"

	"github.com/Atharv-3105/Semantic-Duel/internal/logger"
	"github.com/Atharv-3105/Semantic-Duel/internal/room"
	"github.com/Atharv-3105/Semantic-Duel/internal/semantic"
	"github.com/Atharv-3105/Semantic-Duel/internal/target"
	"github.com/Atharv-3105/Semantic-Duel/internal/ws"
)

type Matchmaker struct {
	queue            []*ws.Client
	mu               sync.Mutex
	rm               *room.Manager
	log              *logger.Logger
	semantic         semantic.Semantic // interface — works with HTTP Client or LocalClient
	targetProvider   *target.Provider
	cleanupCh        chan<- string
	gameDuration     int
	rateLimitSeconds int
}

func New(
	rm *room.Manager,
	log *logger.Logger,
	sc semantic.Semantic,
	targetProvider *target.Provider,
	cleanupCh chan<- string,
	gameDuration int,
	rateLimitSeconds int,
) *Matchmaker {
	return &Matchmaker{
		queue:            make([]*ws.Client, 0),
		rm:               rm,
		log:              log,
		semantic:         sc,
		targetProvider:   targetProvider,
		cleanupCh:        cleanupCh,
		gameDuration:     gameDuration,
		rateLimitSeconds: rateLimitSeconds,
	}
}

// Enqueue adds a client to the matchmaking queue.
// If two players are waiting, they are immediately paired into a new room.
func (m *Matchmaker) Enqueue(client *ws.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.queue = append(m.queue, client)

	if len(m.queue) >= 2 {
		p1 := m.queue[0]
		p2 := m.queue[1]
		m.queue = m.queue[2:]

		m.log.Info("[MATCH] pairing players", "queue_remaining", len(m.queue))

		roomID := fmt.Sprintf("room-%d", time.Now().UnixNano())

		onCleanup := func(id string) {
			m.cleanupCh <- id
		}

		r := room.NewRoom(roomID, p1, p2, m.semantic, onCleanup, m.gameDuration, m.rateLimitSeconds)
		targetWord := m.targetProvider.Random()
		r.Start(targetWord)
		m.rm.Add(r)

		m.log.Info("[MATCH] room created", "room_id", roomID, "target", targetWord)
	}
}