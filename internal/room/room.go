package room

import (
	"log"
	"time"
	"unicode"

	"github.com/Atharv-3105/Graph-Duel/internal/game"
	"github.com/Atharv-3105/Graph-Duel/internal/metrics"
	// "github.com/Atharv-3105/Graph-Duel/internal/matchmaker"
	"github.com/Atharv-3105/Graph-Duel/internal/semantic"

	// "github.com/Atharv-3105/Graph-Duel/internal/semantic"
	"github.com/Atharv-3105/Graph-Duel/internal/ws"
)

type Room struct {
	ID string 
	Player1 *ws.Client
	Player2 *ws.Client
	State   *GameState

	semantic *semantic.Client
	lastSubmit map[string]time.Time

	onCleanup  func(roomID string)
	gameDuration int 	
	rateLimitSeconds int
}


func NewRoom(id string, p1, p2 *ws.Client, sc *semantic.Client, onCleanup func(roomID string), gameDuration int, rateLimitSeconds int) *Room{
	return &Room{
		ID:		id,
		Player1:  p1,
		Player2:  p2,
		semantic: sc,
		lastSubmit:	make(map[string]time.Time),
		onCleanup: onCleanup,
		gameDuration: gameDuration,
	}
}

func (r *Room) Start(target string) {
	r.State = &GameState{
		TargetWord: target,
		Scores:		make(map[string]int),
	}
	// r.lastSubmit = make(map[string]time.Time)
	game.StartGame(&r.State.State)

	metrics.IncGamesStarted()

	//GAME-START Broadcast
	r.broadcast(ws.EventGameStart, ws.GameStartPayload {
		Target: target,
		Duration: r.gameDuration,
	})

	go func() {
		time.Sleep(time.Until(r.State.EndsAt))
		r.endGame()
	}()
}


func (r *Room) UpdateScore(playerID string, similarity float64) {
	score := game.SimilarityToScore(similarity)

	prev, exists := r.State.Scores[playerID]
	if !exists || score > prev {
		r.State.Scores[playerID] = score
	}
}

func (r *Room) HandleWord(playerID, word string) {
	
	//STRICT STATE GATING
	if r.State == nil || r.State.Status != game.Active {
		log.Println("[ROOM] action rejected: game not active", playerID)
		return 
	}
	
	//RATE LIMITING
	now := time.Now()
	last, ok := r.lastSubmit[playerID]
	if ok && now.Sub(last) < time.Second {
		log.Println("[ROOM] rate limit hit", playerID)
		return
	}
	r.lastSubmit[playerID] = now

	//INPUT-VALIDATION
	if len(word) == 0 || len(word) > 32 {
		log.Println("[ROOM] invalid word length", playerID, word)
		return
	}

	for _, ch := range word {
		if !unicode.IsLetter(ch) {
			log.Println("[ROOM] invalid characters", playerID, word)
			return
		}
	}

	//SEMANTIC CALL
	sim, err := r.semantic.Similarity(word, r.State.TargetWord)
	if err != nil {
		log.Println("[ROOM] semantic error", err)
		return 
	}
	
	// if r.State.Status != game.Active{
	// 	return
	// }

	// sim, err := r.semantic.Similarity(word, r.State.TargetWord)
	// if err != nil {
	// 	return 
	// }

	//SCORE NORMALIZATION
	score := game.SimilarityToScore(sim)

	//BEST-SCORE WINS
	prev, exists := r.State.Scores[playerID]
	if !exists || score > prev {
		r.State.Scores[playerID] = score
		log.Println("[ROOM] score updated", playerID, score)
		
		//SCORE-UPDATE BroadCast
		r.broadcast(ws.EventScoreUpdate, ws.ScoreUpdatePayload{
			PlayerID: playerID,
			Score: score,
		})
	}
}

//END-GAME when Disconnection of a Player
func (r *Room) ForceEnd(leaver string) {
	if r.State == nil || r.State.Status == game.Finished{
		return 
	}

	r.State.Status = game.Finished
	log.Println("[ROOM] game ended due to disconnect:", leaver)

	winner := r.Player1.ID
	if leaver == r.Player1.ID {
		winner = r.Player2.ID
	}	

	metrics.IncGamesCompleted()

	//GAME-END Broadcast
	r.broadcast(ws.EventGameOver, ws.GameOverPayload{
		Winner: winner,
		Scores: r.State.Scores,
	})

	r.cleanup()
}

func (r *Room) endGame() {
	if r.State == nil || r.State.Status == game.Finished {
		log.Println("[ROOM] endGame ignored (already finished)")
		return 
	}

	r.State.Status = game.Finished

	winner := ""
	high := -1
	for pid, score := range r.State.Scores{
		if score > high {
			high = score
			winner = pid
		}
	}	

	metrics.IncGamesCompleted()

	r.broadcast(ws.EventGameOver, ws.GameOverPayload{
		Winner: winner, 
		Scores: r.State.Scores,
	})

	r.cleanup()
}


func (r *Room) broadcast(eventType ws.ServerEventType, payload any) {
	r.Player1.Send(eventType, payload)
	r.Player2.Send(eventType, payload)

}


func(r *Room) cleanup() {
	log.Println("[ROOM] cleanup triggered for:", r.ID)
	if r.onCleanup != nil {
		r.onCleanup(r.ID)
	}
}