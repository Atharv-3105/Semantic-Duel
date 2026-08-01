package room

import (
	"context"
	"log"
	"sync"
	"time"
	"unicode"

	"github.com/Atharv-3105/Semantic-Duel/internal/game"
	"github.com/Atharv-3105/Semantic-Duel/internal/metrics"
	"github.com/Atharv-3105/Semantic-Duel/internal/semantic"
	"github.com/Atharv-3105/Semantic-Duel/internal/ws"
)

type Room struct {
	ID      string
	Player1 *ws.Client
	Player2 *ws.Client
	State   *GameState

	mu               sync.Mutex // protects State and lastSubmit
	semantic         semantic.Semantic
	lastSubmit       map[string]time.Time

	onCleanup        func(roomID string)
	gameDuration     int
	rateLimitSeconds int
}

func NewRoom(id string, p1, p2 *ws.Client, sc semantic.Semantic, onCleanup func(roomID string), gameDuration int, rateLimitSeconds int) *Room {
	return &Room{
		ID:               id,
		Player1:          p1,
		Player2:          p2,
		semantic:         sc,
		lastSubmit:       make(map[string]time.Time),
		onCleanup:        onCleanup,
		gameDuration:     gameDuration,
		rateLimitSeconds: rateLimitSeconds,
	}
}

func (r *Room) Start(target string) {
	r.State = &GameState{
		TargetWord: target,
		Scores:     make(map[string]int),
	}
	// Pass gameDuration so the server-side timer matches what the client is shown.
	game.StartGame(&r.State.State, r.gameDuration)

	metrics.IncGamesStarted()

	r.broadcast(ws.EventGameStart, ws.GameStartPayload{
		Target:   target,
		Duration: r.gameDuration,
	})

	go func() {
		time.Sleep(time.Until(r.State.EndsAt))
		r.endGame()
	}()
}

// HandleWord processes a word submission from a player.
//
// The flow is:
//  1. Lock — validate state, rate limit, and input.
//  2. Unlock — release the lock before the ML call so other submissions can proceed.
//  3. Goroutine — call the ML service asynchronously (no more blocking the dispatch goroutine).
//  4. Re-lock — update score and broadcast only if the game is still active.
func (r *Room) HandleWord(playerID, word string) {
	r.mu.Lock()

	// STRICT STATE GATING
	if r.State == nil || r.State.Status != game.Active {
		r.mu.Unlock()
		return
	}

	// RATE LIMITING — uses the configured rateLimitSeconds value
	now := time.Now()
	last, ok := r.lastSubmit[playerID]
	if ok && now.Sub(last) < time.Duration(r.rateLimitSeconds)*time.Second {
		log.Println("[ROOM] rate limit hit:", playerID)
		r.mu.Unlock()
		return
	}
	r.lastSubmit[playerID] = now

	// INPUT VALIDATION
	if len(word) == 0 || len(word) > 32 {
		log.Println("[ROOM] invalid word length:", playerID, word)
		r.mu.Unlock()
		return
	}
	for _, ch := range word {
		if !unicode.IsLetter(ch) {
			log.Println("[ROOM] invalid characters:", playerID, word)
			r.mu.Unlock()
			return
		}
	}

	// Capture the target word before releasing the lock.
	targetWord := r.State.TargetWord
	r.mu.Unlock()

	// ASYNC ML CALL — does not block the message dispatch goroutine.
	go func() {
		sim, err := r.semantic.Similarity(context.Background(), word, targetWord)
		if err != nil {
			log.Println("[ROOM] semantic error:", err)
			return
		}

		score := game.SimilarityToScore(sim)

		r.mu.Lock()
		// Re-check: game may have ended while waiting for the ML response.
		if r.State == nil || r.State.Status != game.Active {
			r.mu.Unlock()
			return
		}

		// BEST-SCORE WINS — only update if this is the player's personal best.
		prev, exists := r.State.Scores[playerID]
		shouldBroadcast := !exists || score > prev
		if shouldBroadcast {
			r.State.Scores[playerID] = score
			log.Println("[ROOM] score updated:", playerID, score)
		}
		r.mu.Unlock()

		// Broadcast outside the lock — Client.Send is safe to call concurrently.
		if shouldBroadcast {
			r.broadcast(ws.EventScoreUpdate, ws.ScoreUpdatePayload{
				PlayerID: playerID,
				Score:    score,
			})
		}
	}()
}

// ForceEnd ends the game immediately because a player disconnected.
// The remaining player is declared the winner.
func (r *Room) ForceEnd(leaver string) {
	r.mu.Lock()
	if r.State == nil || r.State.Status == game.Finished {
		r.mu.Unlock()
		return
	}

	r.State.Status = game.Finished

	winner := r.Player1.ID
	if leaver == r.Player1.ID {
		winner = r.Player2.ID
	}

	scores := copyScores(r.State.Scores)
	r.mu.Unlock()

	log.Println("[ROOM] game force-ended due to disconnect:", leaver)
	metrics.IncGamesCompleted()

	r.broadcast(ws.EventGameOver, ws.GameOverPayload{
		Winner: winner,
		Scores: scores,
	})

	r.cleanup()
}

// endGame is called by the timer goroutine when the game clock reaches zero.
func (r *Room) endGame() {
	r.mu.Lock()
	if r.State == nil || r.State.Status == game.Finished {
		r.mu.Unlock()
		return
	}

	r.State.Status = game.Finished

	// Deterministic winner: explicit comparison instead of relying on
	// non-deterministic map iteration order.
	p1Score := r.State.Scores[r.Player1.ID]
	p2Score := r.State.Scores[r.Player2.ID]
	winner := ""
	switch {
	case p1Score > p2Score:
		winner = r.Player1.ID
	case p2Score > p1Score:
		winner = r.Player2.ID
	// equal scores → winner stays "" which the client interprets as a draw
	}

	scores := copyScores(r.State.Scores)
	r.mu.Unlock()

	log.Println("[ROOM] game ended normally, winner:", winner)
	metrics.IncGamesCompleted()

	r.broadcast(ws.EventGameOver, ws.GameOverPayload{
		Winner: winner,
		Scores: scores,
	})

	r.cleanup()
}

// broadcast sends an event to both players.
// Safe to call without holding r.mu — Player1/Player2 never change after NewRoom.
func (r *Room) broadcast(eventType ws.ServerEventType, payload any) {
	r.Player1.Send(eventType, payload)
	r.Player2.Send(eventType, payload)
}

func (r *Room) cleanup() {
	log.Println("[ROOM] cleanup triggered for:", r.ID)
	if r.onCleanup != nil {
		r.onCleanup(r.ID)
	}
}

// copyScores returns a shallow copy of the scores map so the caller
// can safely read it after releasing the mutex.
func copyScores(src map[string]int) map[string]int {
	dst := make(map[string]int, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}