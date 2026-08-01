package matchmaker

import (
	"context"
	"testing"

	"github.com/Atharv-3105/Semantic-Duel/internal/logger"
	"github.com/Atharv-3105/Semantic-Duel/internal/room"
	"github.com/Atharv-3105/Semantic-Duel/internal/target"
	"github.com/Atharv-3105/Semantic-Duel/internal/ws"
)

// mockSemantic satisfies semantic.Semantic without making any HTTP calls.
type mockSemantic struct{}

func (m *mockSemantic) Similarity(_ context.Context, _, _ string) (float64, error) {
	return 0.5, nil
}

// newTestMatchmaker creates a Matchmaker wired to real (but in-memory) dependencies.
// gameDuration is set to 3600s so the timer goroutine spawned by room.Start()
// doesn't fire during tests.
func newTestMatchmaker() (*Matchmaker, *room.Manager, chan string) {
	log := logger.New()
	rm := room.NewManager(log)
	sc := &mockSemantic{}
	tp := target.New(target.DefaultWords)
	cleanupCh := make(chan string, 16)
	mm := New(rm, log, sc, tp, cleanupCh, 3600, 0)
	return mm, rm, cleanupCh
}

// newTestClient creates a ws.Client with no real WebSocket connection.
func newTestClient(id string) *ws.Client {
	c := ws.NewClient(nil)
	c.ID = id
	return c
}

// ─── Queue behaviour ─────────────────────────────────────────────────────────

func TestMatchmaker_SingleClientDoesNotCreateRoom(t *testing.T) {
	mm, rm, _ := newTestMatchmaker()

	p1 := newTestClient("player-1")
	mm.Enqueue(p1)

	if _, ok := rm.RoomForClient("player-1"); ok {
		t.Error("a room should not be created with only one player in the queue")
	}
}

func TestMatchmaker_TwoClientsTriggerRoomCreation(t *testing.T) {
	mm, rm, _ := newTestMatchmaker()

	p1 := newTestClient("player-1")
	p2 := newTestClient("player-2")

	mm.Enqueue(p1)
	mm.Enqueue(p2)

	if _, ok := rm.RoomForClient("player-1"); !ok {
		t.Error("player-1 should be assigned to a room after two enqueues")
	}
	if _, ok := rm.RoomForClient("player-2"); !ok {
		t.Error("player-2 should be assigned to the same room")
	}
}

func TestMatchmaker_BothPlayersInSameRoom(t *testing.T) {
	mm, rm, _ := newTestMatchmaker()

	p1 := newTestClient("player-1")
	p2 := newTestClient("player-2")

	mm.Enqueue(p1)
	mm.Enqueue(p2)

	r1, ok1 := rm.RoomForClient("player-1")
	r2, ok2 := rm.RoomForClient("player-2")

	if !ok1 || !ok2 {
		t.Fatal("both players should have a room assigned")
	}
	if r1 != r2 {
		t.Error("both players should be in the same room instance")
	}
}

func TestMatchmaker_ThirdClientWaitsAfterPairing(t *testing.T) {
	mm, rm, _ := newTestMatchmaker()

	p1 := newTestClient("player-1")
	p2 := newTestClient("player-2")
	p3 := newTestClient("player-3")

	mm.Enqueue(p1)
	mm.Enqueue(p2) // p1+p2 are paired
	mm.Enqueue(p3) // p3 waits alone

	if _, ok := rm.RoomForClient("player-3"); ok {
		t.Error("player-3 should still be waiting — no fourth player has joined")
	}
}

func TestMatchmaker_FourClientsCreateTwoRooms(t *testing.T) {
	mm, rm, _ := newTestMatchmaker()

	clients := []*ws.Client{
		newTestClient("p1"),
		newTestClient("p2"),
		newTestClient("p3"),
		newTestClient("p4"),
	}

	for _, c := range clients {
		mm.Enqueue(c)
	}

	for _, c := range clients {
		if _, ok := rm.RoomForClient(c.ID); !ok {
			t.Errorf("player %s should be in a room after four enqueues", c.ID)
		}
	}

	// The two rooms must be different.
	r12, _ := rm.RoomForClient("p1")
	r34, _ := rm.RoomForClient("p3")
	if r12 == r34 {
		t.Error("four players should be split into two separate rooms")
	}
}

func TestMatchmaker_QueueIsEmptyAfterPairing(t *testing.T) {
	mm, _, _ := newTestMatchmaker()

	p1 := newTestClient("player-1")
	p2 := newTestClient("player-2")

	mm.Enqueue(p1)
	mm.Enqueue(p2)

	// Internal queue should now be drained (length 0).
	mm.mu.Lock()
	qLen := len(mm.queue)
	mm.mu.Unlock()

	if qLen != 0 {
		t.Errorf("expected empty queue after pairing, got length %d", qLen)
	}
}

func TestMatchmaker_QueueLengthOneAfterOddEnqueues(t *testing.T) {
	mm, _, _ := newTestMatchmaker()

	for i := range 5 {
		c := newTestClient("player-" + string(rune('A'+i)))
		mm.Enqueue(c)
	}

	// 5 enqueues → 2 pairs created, 1 left waiting.
	mm.mu.Lock()
	qLen := len(mm.queue)
	mm.mu.Unlock()

	if qLen != 1 {
		t.Errorf("expected queue length 1 after 5 enqueues, got %d", qLen)
	}
}
