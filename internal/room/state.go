package room

import (
	"github.com/Atharv-3105/Semantic-Duel/internal/game"
)

type GameState struct {
	TargetWord string
	Scores     map[string]int
	game.State
}
